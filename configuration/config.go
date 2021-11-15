package configuration

import (
	"fmt"
	"reflect"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/2gis/loggo/common"
)

// Parallelism is a count of workers in each stage to run
const Parallelism = 8

type K8SExtends struct {
	DataCenter     string
	Purpose        string
	NodeHostname   string
	LogType        string
	LogstashPrefix string
}

func (e *K8SExtends) EntryMap() common.EntryMap {
	return common.EntryMap{
		common.KubernetesNodeHostname: e.NodeHostname,
		common.LabelDataCenter:        e.DataCenter,
		common.LabelPurpose:           e.Purpose,
		common.LabelLogstashPrefix:    e.LogstashPrefix,
		common.LabelLogType:           e.LogType,
	}
}

type FollowerConfig struct {
	ReaderBufferSize                  int
	CursorCommitIntervalSec           int
	NoRecordsSleepIntervalSec         int
	ThrottlingLimitsUpdateIntervalSec int
	FromTailFlag                      bool
}

type ParserConfig struct {
	UserLogFieldsKey string
	CRIFieldsKey     string
	ExtendsFieldsKey string
	RawLogFieldKey   string

	FlattenUserLog bool
}

type JournaldConfig struct {
	LogJournalD   bool
	JournaldPath  string
	MachineIDPath string
}

type SLIExporterConfig struct {
	Enabled bool

	K8SConfigPath        string
	Buckets              string
	ServiceSourcePath    string
	ServiceDefaultDomain string

	ServiceUpdateIntervalSec int

	AnnotationExporterEnable string
	AnnotationExporterPaths  string
	AnnotationSLADomains     string
}

type RedisTransportConfig struct {
	URL      string
	Password string
	Key      string
}
type AMQPTransportConfig struct {
	URL      string
	Exchange string
	Key      string
}

type FirehoseTransportConfig struct {
	DeliveryStream string
}

// Config stores all configuration from environment or launch keys
type Config struct {
	K8SExtends              K8SExtends
	ParserConfig            ParserConfig
	FollowerConfig          FollowerConfig
	JournaldConfig          JournaldConfig
	SLIExporterConfig       SLIExporterConfig
	FirehostTransportConfig FirehoseTransportConfig
	AMQPTransportConfig     AMQPTransportConfig
	RedisTransportConfig    RedisTransportConfig

	TransportBufferSizeMax int
	Transport              string

	LogsPath                 string
	PositionFilePath         string
	ContainersIgnoreFilePath string

	ReadRateRulesPath string
	ReadRateDefault   float64

	TargetsRefreshIntervalSec int
	FlushIntervalSec          int
	MetricsResetIntervalSec   int

	LogLevel  string
	LogFormat string
}

// GetConfig generates Config from options and env vars
func GetConfig() Config {
	config := Config{
		K8SExtends:     K8SExtends{},
		FollowerConfig: FollowerConfig{},
		JournaldConfig: JournaldConfig{},
	}

	// common
	kingpin.Flag("position-file-path", "Path to file where loggo stores log file cursors").
		Default("/var/log/loggo-logs.pos").
		Envar("POSITION_FILE_PATH").
		StringVar(&config.PositionFilePath)
	kingpin.Flag("containers-ignore-file-path", "Path to file where loggo stores ignored containers IDs").
		Default("/var/log/loggo-containers-ignore").
		Envar("CONTAINERS_IGNORE_FILE_PATH").
		StringVar(&config.ContainersIgnoreFilePath)
	kingpin.Flag("logs-path", "Path where loggo will watch for log files").
		Default("/var/log/pods/").
		Envar("LOGS_PATH").
		StringVar(&config.LogsPath)
	kingpin.Flag(
		"targets-refresh-interval-sec",
		"How often reread logs-path directory searching for new log files").
		Default("10").
		Envar("TARGETS_REFRESH_INTERVAL_SEC").
		IntVar(&config.TargetsRefreshIntervalSec)

	// transport
	kingpin.Flag("transport", "Transport type for log messages [amqp | redis]").
		Default("amqp").
		Envar("TRANSPORT").
		StringVar(&config.Transport)
	kingpin.Flag("redis-hostname", "Redis host URL to use.").
		Default("localhost:6379").
		Envar("REDIS_HOSTNAME").
		StringVar(&config.RedisTransportConfig.URL)
	kingpin.Flag("redis-password", "Redis password to use.").
		Envar("REDIS_PASSWORD").
		StringVar(&config.RedisTransportConfig.Password)
	kingpin.Flag("redis-key", "Key of a list in Redis where to send messages.").
		Default("k8s-logs").
		Envar("REDIS_KEY").
		StringVar(&config.RedisTransportConfig.Key)
	kingpin.Flag("amqp-url", "AMQP host URL to use.").
		Default("amqp://localhost/").
		Envar("AMQP_URL").
		StringVar(&config.AMQPTransportConfig.URL)
	kingpin.Flag("amqp-exchange", "AMQP Exchange for log message delivery").
		Default("amq.direct").
		Envar("AMQP_EXCHANGE").
		StringVar(&config.AMQPTransportConfig.Exchange)
	kingpin.Flag("amqp-routing-key", "AMQP routing key for message delivery").
		Default("all-other").
		Envar("AMQP_ROUTING_KEY").
		StringVar(&config.AMQPTransportConfig.Key)
	kingpin.Flag("firehose-delivery-stream", "AWS Firehose delivery stream.").
		Default("default-delivery").
		Envar("FIREHOSE_DELIVERY_STREAM").
		StringVar(&config.FirehostTransportConfig.DeliveryStream)
	kingpin.Flag("flush-interval-sec", "How often to try sending data to transport").
		Default("60").
		Envar("FLUSH_INTERVAL_SEC").
		IntVar(&config.FlushIntervalSec)
	kingpin.Flag("buffer-max-size", "Maximum log messages to buffer before sending to storage").
		Default("1000").
		Envar("BUFFER_SIZE_MAX").
		IntVar(&config.TransportBufferSizeMax)

	// readers
	kingpin.Flag(
		"reader-buffer-size",
		"Size of readers internal buffer, bytes; hence, the maximum log message length").
		Default("32000").
		Envar("READER_BUFFER_SIZE").
		IntVar(&config.FollowerConfig.ReaderBufferSize)
	kingpin.Flag("throttling-limits-update-interval-sec", "How often to get updates from throttling config").
		Default("600").
		Envar("THROTTLING_LIMITS_UPDATE_INTERVAL_SEC").
		IntVar(&config.FollowerConfig.ThrottlingLimitsUpdateIntervalSec)
	kingpin.Flag(
		"no-records-sleep-sec",
		"How long to wait before start reading logfile which hadn't logs added recently").
		Default("4").
		Envar("NO_RECORDS_SLEEP_SEC").
		IntVar(&config.FollowerConfig.NoRecordsSleepIntervalSec)
	kingpin.Flag("cursor-commit-interval-sec", "How often to try sending data to transport").
		Default("60").
		Envar("CURSOR_COMMIT_INTERVAL_SEC").
		IntVar(&config.FollowerConfig.CursorCommitIntervalSec)
	kingpin.Flag("from-tail", "Whether to start reading files without valid cursors from tail, default false").
		Default("false").
		Envar("FROM_TAIL_FLAG").
		BoolVar(&config.FollowerConfig.FromTailFlag)

	// system journal reader
	kingpin.Flag("log-journald", "Whether to log journald or not, default true").
		Default("true").
		Envar("LOG_JOURNALD").
		BoolVar(&config.JournaldConfig.LogJournalD)
	kingpin.Flag("machine-id-path", "Path to file with machine identifier").
		Default("/etc/machine-id").
		Envar("MACHINE_ID_PATH").
		StringVar(&config.JournaldConfig.MachineIDPath)
	kingpin.Flag("journald-path", "Journald journal directory path").
		Default("/var/log/journal/").
		Envar("JOURNALD_PATH").
		StringVar(&config.JournaldConfig.JournaldPath)

	// throttling
	kingpin.Flag(
		"read-rate-rules-path",
		"Path to file with throttling rules. If not specified, READ_RATE_DEFAULT will be used for all containers").
		Default("").
		Envar("READ_RATE_RULES_PATH").
		StringVar(&config.ReadRateRulesPath)
	kingpin.Flag(
		"read-rate-default",
		"Maximum log messages per container to read per second by default. Must be greater than zero.").
		Default("1000").
		Envar("READ_RATE_DEFAULT").
		Float64Var(&config.ReadRateDefault)

	// extends
	kingpin.Flag("dc", "Current datacenter id, will be included to each log message").
		Default("n3").
		Envar("DC").
		StringVar(&config.K8SExtends.DataCenter)
	kingpin.Flag("purpose", "Current datacenter purpose, will be included to each log message").
		Default("staging").
		Envar("PURPOSE").
		StringVar(&config.K8SExtends.Purpose)
	kingpin.Flag("node-hostname", "Current node hostname, will be included to each log message").
		Default("localhost").
		Envar("NODE_HOSTNAME").
		StringVar(&config.K8SExtends.NodeHostname)
	kingpin.Flag("log-type", "Current log type, will be included to each log message").
		Default("containers").
		Envar("LOG_TYPE").
		StringVar(&config.K8SExtends.LogType)
	kingpin.Flag("logstash-prefix", "Current logstash prefix, will be included to each log message").
		Default("k8s-unknown").
		Envar("LOGSTASH_PREFIX").
		StringVar(&config.K8SExtends.LogstashPrefix)

	// sla exporter
	kingpin.Flag("sla-exporter", "Whether to export SLA or not, default true").
		Default("true").
		Envar("SLA_EXPORTER").
		BoolVar(&config.SLIExporterConfig.Enabled)
	kingpin.Flag("k8s-config-path", "K8s config file path. Not required in current configuration").
		Envar("K8S_CONFIG_PATH").
		StringVar(&config.SLIExporterConfig.K8SConfigPath)
	kingpin.Flag("sla-service-source-path", "Enables SLA in non-K8S mode; path to services declaration").
		Default("").
		Envar("SLA_SERVICE_SOURCE_PATH").
		StringVar(&config.SLIExporterConfig.ServiceSourcePath)
	kingpin.Flag("default-service-domain", "").
		Default("2gis.test").
		Envar("SERVICE_DEFAULT_DOMAIN").
		StringVar(&config.SLIExporterConfig.ServiceDefaultDomain)
	kingpin.Flag("service-update-interval-sec", "How often to get updates from K8S Api Server").
		Default("60").
		Envar("SERVICE_UPDATE_INTERVAL_SEC").
		IntVar(&config.SLIExporterConfig.ServiceUpdateIntervalSec)
	kingpin.Flag("sla-buckets", "Space-delimited float values of histogram bucket upper borders to use").
		Default("0.01 0.02 0.04 0.06 0.08 0.1 0.15 0.2 0.25 0.3 0.4 0.5 0.6 0.7 0.8 0.9 1 1.2 1.5 1.75 2 3 4 5 8 10 20 60").
		Envar("SLA_BUCKETS").
		StringVar(&config.SLIExporterConfig.Buckets)
	// backward compatibility with annotations
	kingpin.Flag("sla-annotation-enable", "K8S service default enable annotation rewrite").
		Default(AnnotationExporterEnableDefault).
		Envar("SLA_SERVICE_ANNOTATION_ENABLE").
		StringVar(&config.SLIExporterConfig.AnnotationExporterEnable)
	kingpin.Flag("sla-annotation-paths", "K8S service default paths annotation rewrite").
		Default(AnnotationExporterPathsDefault).
		Envar("SLA_SERVICE_ANNOTATION_PATHS").
		StringVar(&config.SLIExporterConfig.AnnotationExporterPaths)
	kingpin.Flag("sla-annotation-domains", "K8S service default domains annotation rewrite").
		Default(AnnotationSLADomainsDefault).
		Envar("SLA_SERVICE_ANNOTATION_DOMAINS").
		StringVar(&config.SLIExporterConfig.AnnotationSLADomains)

	kingpin.Flag("user-log-fields-key", "Entry field where user log should be put.").
		Default("").
		Envar("USER_LOG_FIELDS_KEY").
		StringVar(&config.ParserConfig.UserLogFieldsKey)

	kingpin.Flag("cri-fields-key", "Entry field where docker/containerd engine fields map should be put.").
		Default("").
		Envar("cri_FIELDS_KEY").
		StringVar(&config.ParserConfig.CRIFieldsKey)

	kingpin.Flag("extends-fields-key", "Entry field where loggo and k8s extends fields map should be put.").
		Default("").
		Envar("EXTENDS_FIELDS_KEY").
		StringVar(&config.ParserConfig.ExtendsFieldsKey)
	kingpin.Flag("raw-log-field-key", "Entry field inside user log fields map. Used for non-json messages.").
		Default("msg").
		Envar("RAW_LOG_FIELD_KEY").
		StringVar(&config.ParserConfig.RawLogFieldKey)

	kingpin.Flag("flatten-user-log", "Whether to flatten user log or not.").
		Default("true").
		Envar("FLATTEN_USER_LOG").
		BoolVar(&config.ParserConfig.FlattenUserLog)

	kingpin.Flag("metrics-reset-interval-sec", "Prometheus metrics reset interval.").
		Default("172800").
		Envar("METRICS_RESET_INTERVAL_SEC").
		IntVar(&config.MetricsResetIntervalSec)

	// logging
	kingpin.Flag("log-level", "Loggo main log level").
		Default("warning").
		Envar("LOG_LEVEL").
		StringVar(&config.LogLevel)
	kingpin.Flag("log-format", "Loggo main log format").
		Default("json").
		Envar("LOG_FORMAT").
		StringVar(&config.LogFormat)

	kingpin.Parse()
	return config
}

// ToString converts config to table formatted multiline string
func (config *Config) ToString() string {
	v := reflect.ValueOf(*config)
	output := "\n"

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			output += fmt.Sprintf("%s:\t'%v'\n", v.Type().Field(i).Name, v.Field(i).Interface())
		}
	}
	return output
}
