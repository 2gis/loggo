package transport

/* transport types */
const (
	TypeAMQP     = "amqp"
	TypeRedis    = "redis"
	TypeFirehose = "firehose"
)

// RedisMaxIdleConnections default.
const RedisMaxIdleConnections = 100

var TypesSupported = []string{TypeRedis, TypeAMQP, TypeFirehose}
