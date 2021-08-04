package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/prometheus/pkg/textparse"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/2gis/loggo/transport/amqpclient"
	"github.com/2gis/loggo/transport/redisclient"
)

const (
	metricsURL       = "http://localhost:8080/metrics"
	connectRetryMax  = 15
	sentRecordsCount = 20
)

type MetricMatch struct {
	metric    string
	value     float64
	foundFlag bool
}

type nginxLog struct {
	TimeMSec  string `json:"time_msec"`
	RequestID string `json:"request_id"`
	foundFlag bool
}

type config struct {
	amqpURL        string
	amqpExchange   string
	amqpRoutingKey string
	redisURL       string
	redisKey       string
	transport      string

	createRabbitQueues bool
	slaTesting         bool
}

func main() {
	config := config{}
	kingpin.Flag("transport", "Transport type for log messages [amqp | redis]").
		Default("amqp").
		Envar("TRANSPORT").
		StringVar(&config.transport)
	kingpin.Flag("sla-testing", "Enable gathering of metrics").
		Envar("SLA_TESTING").
		BoolVar(&config.slaTesting)
	kingpin.Flag("create-rabbit-queues", "Create rabbit queues for testing").
		Envar("CREATE_RABBIT_QUEUES").
		BoolVar(&config.createRabbitQueues)
	kingpin.Flag("amqp-url", "Where to send log messages").
		Default("amqp://localhost/").
		Envar("AMQP_URL").
		StringVar(&config.amqpURL)
	kingpin.Flag("amqp-exchange", "AMQP Exchange for log message delivery").
		Default("amq.direct").
		Envar("AMQP_EXCHANGE").
		StringVar(&config.amqpExchange)
	kingpin.Flag("amqp-routing-key", "AMQP routing key for message delivery").
		Default("all-other").
		Envar("AMQP_ROUTING_KEY").
		StringVar(&config.amqpRoutingKey)
	kingpin.Flag("redis-url", "Where to send log messages").
		Default("localhost:6379").
		Envar("REDIS_URL").
		StringVar(&config.redisURL)
	kingpin.Flag("redis-key", "Redis key for message delivery").
		Default("k8s-logs").
		Envar("REDIS_KEY").
		StringVar(&config.redisKey)
	kingpin.Parse()

	if config.slaTesting {
		testMetrics()
		return
	}

	if config.transport == "amqp" {
		if config.createRabbitQueues {
			createRabbitQueues(config, 30, 1)
			return
		}

		testAmqp(config)
		return

	}

	// else it's redis
	testRedis(config)
}

func testMetrics() {
	var iteration int
	var metricsResponse *http.Response
	var err error

	// establish connection to and get data from metrics endpoint
	for {
		metricsResponse, err = http.Get(metricsURL)

		if err == nil {
			break
		}

		if iteration > connectRetryMax {
			log.Fatal(err)
		}

		time.Sleep(1 * time.Second)
		iteration++
	}

	defer metricsResponse.Body.Close()
	var metrics []byte

	if metricsResponse.StatusCode != http.StatusOK {
		log.Fatalf("Response from %s, not 200. %d", metricsURL, metricsResponse.StatusCode)
	}

	metrics, err = ioutil.ReadAll(metricsResponse.Body)

	if err != nil {
		log.Fatalf("Unable to read metrics from reponse, %s", err.Error())
	}

	// parse metrics and compare it with expectations
	expected := getMetricsExpectations()
	parser := textparse.New(metrics)

	for parser.Next() {
		metric, _, value := parser.At()
		for _, element := range expected {
			if string(metric) == element.metric && value == element.value {
				element.foundFlag = true
			}
		}
	}

	passedFlag := true

	for _, match := range expected {
		if !match.foundFlag {
			passedFlag = false
			log.Printf("Error: Expected metric not found '%s=%f'", match.metric, match.value)
		}
	}

	if !passedFlag {
		log.Println(string(metrics))
		log.Fatalln("SLA tests failed, see errors above")
	}

	log.Println("SLA tests ok.")
}

func testAmqp(config config) {
	// establish connection to and get data from broker
	broker, err := amqpclient.NewAMQPClient(config.amqpURL, config.amqpExchange, config.amqpRoutingKey)

	if err != nil {
		log.Fatalf("Unable to init amqp client. %s", err)
	}

	defer broker.Close()

	messages, err := broker.Consume("test")

	if err != nil {
		log.Fatalf("Consumption failed: %s", err)
	}

	var messageCount int
	expectations := getLogExpectations()

	log.Println("Waiting for messages in channel")

	// parse messages and compare it with expectations
	for message := range messages {
		received := &nginxLog{}
		err := json.Unmarshal(message.Body, received)

		if err != nil {
			log.Printf("Unable to parse %s, %s", message.Body, err)
		}

		checkMatchExpected(received, expectations)

		messageCount++
		if messageCount < sentRecordsCount {
			continue
		}

		break
	}

	for _, expectation := range expectations {
		if !expectation.foundFlag {
			log.Fatalf("Test failed, not all log files matched")
		}
	}

	log.Println("AMQP transmission tests ok")
}

func testRedis(c config) {
	// establish connection to and get data from broker
	client := redisclient.NewRedisClient(c.redisURL, c.redisKey)
	defer client.Close()

	expectations := getLogExpectations()
	log.Println("Waiting for messages in channel")

	// parse messages and compare it with expectations
	for i := 0; i < sentRecordsCount; i++ {
		message, err := client.ReceiveMessage()

		if err != nil {
			log.Fatalf("Receive failed: %s", err)
		}

		l := &nginxLog{}
		err = json.Unmarshal(message, l)

		if err != nil {
			log.Printf("Unable to parse %s, %s", message, err)
		}

		checkMatchExpected(l, expectations)
	}

	for i, expectation := range expectations {
		if !expectation.foundFlag {
			log.Println(i)
		}
	}

	for _, expectation := range expectations {
		if !expectation.foundFlag {
			log.Fatalf("Test failed, not all log files matched")
		}
	}

	log.Println("Redis transmission tests ok")
}

func checkMatchExpected(actual *nginxLog, expected []*nginxLog) {
	for _, expectation := range expected {
		if !(expectation.TimeMSec == actual.TimeMSec && expectation.RequestID == actual.RequestID) {
			continue
		}

		expectation.foundFlag = true
	}
}

func createRabbitQueues(config config, retries int, timeout int) {
	var broker *amqpclient.Broker
	var err error

	for i := 0; i < retries; i++ {
		broker, err = amqpclient.NewAMQPClient(config.amqpURL, config.amqpExchange, config.amqpRoutingKey)
		if err != nil {
			log.Printf("Try #%d, Unable to init amqp client. %s, retry after timeout %d", i, err, timeout)
			time.Sleep(time.Duration(timeout) * time.Second)
			continue
		}
		log.Println("Connection to amqp initialized.")
		break
	}
	defer broker.Close()
	_ = broker.CreateExchange()
	broker.CreateQueue("test", true)
	_ = broker.BindQueue("test")
}

func getLogExpectations() []*nginxLog {
	return []*nginxLog{
		{TimeMSec: "1515474468.380", RequestID: "e7bfe08ae0d17cf3d8ba83fea81a672b"},
		{TimeMSec: "1515474468.381", RequestID: "551c1c38c45f5086260a5ba919cab365"},
		{TimeMSec: "1515474468.396", RequestID: "4a1038746802c61a1de10c8802b850c8"},
		{TimeMSec: "1515474468.396", RequestID: "d010f6631c81407596a7ff19aaf6312b"},
		{TimeMSec: "1515474468.396", RequestID: "bb994be4fd534ffc4e9605b70c18ee9c"},
		{TimeMSec: "1515474468.396", RequestID: "a9a150d1c4ab46b15be72da07f6ae35a"},
		{TimeMSec: "1515474468.757", RequestID: "9ae268549cac561a97678a69431868c8"},
		{TimeMSec: "1515474468.821", RequestID: "9ae268549cac561a97678a69431868c8"},
		{TimeMSec: "1515474468.942", RequestID: "d44ee9cad42d204e9dad7acff8cc5466"},
		{TimeMSec: "1515474469.061", RequestID: "5edc74969d170dd61eaf1cfa90f4ebad"},
		{TimeMSec: "1515474469.526", RequestID: "b6f12e8ded88b2fa6dc598f5fca1c978"},
		{TimeMSec: "1515474469.561", RequestID: "b6f12e8ded88b2fa6dc598f5fca1c978"},
		{TimeMSec: "1515474469.761", RequestID: "85c9dc2c93541d5163b4092ee8a449fb"},
		{TimeMSec: "1515474469.812", RequestID: "285e18e863aa016bb975f08d0761abc6"},
		{TimeMSec: "1515474470.067", RequestID: "23b2aa35cc21972124e9801d6657d4b6"},
		{TimeMSec: "1515474470.215", RequestID: "9ea29074902088fc11f4f21065abf177"},
		{TimeMSec: "1515474470.251", RequestID: "1216ed217cb8c6c0d76d0da311c91b26"},
	}
}

func getMetricsExpectations() []*MetricMatch {
	return []*MetricMatch{
		{
			metric: `http_request_count{method="POST",path="all",protocol="HTTP/1.1",service="B",status="200",upstream_pod_name=""}`,
			value:  1.0,
		},
		{
			metric: `http_request_time_sum{method="POST",path="all",protocol="HTTP/1.1",service="B",upstream_pod_name=""}`,
			value:  0.01,
		},
		{
			metric: `http_request_count{method="POST",path="all",protocol="HTTP/1.1",service="A",status="404",upstream_pod_name=""}`,
			value:  1.0,
		},
		{
			metric: `http_request_time_sum{method="POST",path="all",protocol="HTTP/1.1",service="A",upstream_pod_name=""}`,
			value:  0.005,
		},
		{
			metric: `http_upstream_response_time_total_sum{method="POST",path="all",protocol="HTTP/1.1",service="B",upstream_pod_name=""}`,
			value:  0.098,
		},
		{
			metric: `http_upstream_response_time_total_count{method="POST",path="all",protocol="HTTP/1.1",service="A",upstream_pod_name=""}`,
			value:  1.0,
		},
	}
}
