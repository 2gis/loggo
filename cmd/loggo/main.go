package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/2gis/loggo/components"

	"github.com/2gis/loggo/common"
	"github.com/2gis/loggo/components/containers"
	"github.com/2gis/loggo/dispatcher"
	"github.com/2gis/loggo/parsers"
	"github.com/2gis/loggo/stages"

	"github.com/2gis/loggo/components/k8s"
	"github.com/2gis/loggo/components/rates"
	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/dispatcher/workers"
	"github.com/2gis/loggo/logging"
	"github.com/2gis/loggo/metrics"
	"github.com/2gis/loggo/storage"
	"github.com/2gis/loggo/transport"
	"github.com/2gis/loggo/transport/amqpclient"
	"github.com/2gis/loggo/transport/firehoseclient"
	"github.com/2gis/loggo/transport/redisclient"
)

func main() {
	config := configuration.GetConfig()
	log.Printf("Starting with configuration: %s", config.ToString())
	logger := logging.NewLogger("json", config.LogLevel, os.Stdout)

	var transportClient transport.Client
	var err error

	switch config.Transport {
	case transport.TypeAMQP:
		transportClient, err = amqpclient.NewAMQPClient(config.AMQPTransportConfig)

		if err != nil {
			logger.Fatalf("Unable to init amqp client, %s", err)
		}
	case transport.TypeRedis:
		transportClient = redisclient.NewRedisClient(config.RedisTransportConfig)
	case transport.TypeFirehose:
		transportClient, err = firehoseclient.NewFireHoseClient(config.FirehostTransportConfig)

		if err != nil {
			logger.Fatalf("Unable to init firehose client, %s", err)
		}
	default:
		logger.Fatalf(
			"Unsupported transport type, supported types: [%s].",
			strings.Join(transport.TypesSupported, ", "),
		)
	}

	cursorStorage, err := storage.NewStorage(config.PositionFilePath, 1)
	if err != nil {
		logger.Fatalln(err)
	}

	providerContainers, err := containers.NewProviderContainers(config.LogsPath, logger)
	if err != nil {
		logger.Fatalln(err)
	}

	var providerK8SServices k8s.ServicesProvider = k8s.NewProviderStub()

	if config.SLIExporterConfig.Enabled {
		switch config.SLIExporterConfig.ServiceSourcePath {
		case "":
			k8sConfig, err := configuration.K8sConfig(config.SLIExporterConfig.K8SConfigPath)
			if err != nil {
				logger.Fatal(err)
			}

			logger.Printf("Info: Init Kubernetes client with APIPath %s", k8sConfig.APIPath)

			client, err := kubernetes.NewForConfig(k8sConfig)
			if err != nil {
				logger.Fatal(err)
			}

			providerK8SServices = k8s.NewProviderK8SServices(client, config.SLIExporterConfig, logger)
		default:
			providerFile, err := k8s.NewProviderFile(config.SLIExporterConfig)

			if err != nil {
				logger.Fatal(err)
			}

			providerK8SServices = providerFile
		}
	}

	metricsCollector, err := metrics.NewCollector(config.SLIExporterConfig.Buckets)
	if err != nil {
		logger.Fatalln(err)
	}

	var recordsProvider rates.RateRecordsProvider = rates.NewRuleRecordsProviderStub()

	if len(config.ReadRateRulesPath) != 0 {
		recordsProvider = rates.NewRateRecordsProviderYaml(config.ReadRateRulesPath)
	}

	rater, err := rates.NewRater(recordsProvider, config.ReadRateDefault)
	if err != nil {
		logger.Fatalln(err)
	}

	var parserSLI stages.ParserSLI = parsers.NewParserSliStub()

	if config.SLIExporterConfig.Enabled {
		parserSLI = parsers.NewParserSLI(providerK8SServices, metricsCollector)
	}

	go metrics.ServeHTTPRequests(":8080", "/metrics")

	transportInputs := make([]<-chan string, 0, 2)
	wg := &sync.WaitGroup{}
	ctx, stop := context.WithCancel(context.Background())

	go components.RetrievePeriodic(
		ctx,
		providerK8SServices,
		time.Duration(config.SLIExporterConfig.ServiceUpdateIntervalSec)*time.Second,
		logger,
	)
	go components.RetrievePeriodic(
		ctx,
		metricsCollector,
		time.Duration(config.MetricsResetIntervalSec)*time.Second,
		logger,
	)
	go components.RetrievePeriodic(
		ctx,
		rater,
		time.Duration(config.FollowerConfig.ThrottlingLimitsUpdateIntervalSec)*time.Second,
		logger,
	)

	followerFabric := workers.NewFollowersFabric(
		config,
		metricsCollector,
		cursorStorage,
		rater,
		logger,
	)
	workersDispatcher := dispatcher.NewDispatcher(
		followerFabric,
		providerContainers,
		cursorStorage,
		config.JournaldConfig.LogJournalD,
		time.Duration(config.TargetsRefreshIntervalSec)*time.Second,
		logger,
	)
	transportInputs = append(transportInputs, workersDispatcher.OutJournald())

	wg.Add(1)
	go func() {
		defer wg.Done()
		workersDispatcher.Start(ctx)
	}()

	stageParsing := stages.NewStageParsingEntry(
		workersDispatcher.Out(),
		parsers.CreateParserDockerFormat(config.ParserConfig),
		parsers.CreateParserContainerDFormat(config.ParserConfig),
		parsers.ParseStringPlain,
		config.ParserConfig.ExtendsFieldsKey,
		logger,
	)

	stageParsingSLI := stages.NewStageParsingSLI(
		stageParsing.Out(),
		parserSLI,
		logger,
	)
	stageFiltering := stages.NewStageFiltering(
		stageParsingSLI.Out(),
		config.ParserConfig.UserLogFieldsKey,
		logger,
	)
	stageMarshalling := stages.NewStageJSONMarshalling(
		stageFiltering.Out(),
		logger,
	)
	transportInputs = append(transportInputs, stageMarshalling.Out())

	stageTransport := stages.NewStageTransport(
		common.MergeChannelsString(transportInputs...),
		transportClient,
		config.TransportBufferSizeMax,
		time.Duration(config.FlushIntervalSec)*time.Second,
		logger,
	)

	for _, stage := range []stages.Stage{
		stageParsing, stageParsingSLI, stageFiltering, stageMarshalling, stageTransport} {
		wg.Add(1)

		go func(stage stages.Stage) {
			defer wg.Done()
			stages.StageInit(stage, configuration.Parallelism)
		}(stage)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	logger.Printf("Catched signal '%s'", <-ch)

	stop()
	wg.Wait()

	if transportClient.Close() != nil {
		logger.Errorf("workersDispatcher: error during closing the transport: '%s'", err)
	}

	if err = cursorStorage.Close(); err != nil {
		logger.Error(err)
	}

	logger.Println("Loggo has been stopped.")
}
