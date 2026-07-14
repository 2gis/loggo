package transport

/* transport types */
const (
	TypeAMQP     = "amqp"
	TypeRedis    = "redis"
	TypeFirehose = "firehose"
	TypeNoop     = "noop"
)

// RedisMaxIdleConnections default.
const RedisMaxIdleConnections = 100

var TypesSupported = []string{TypeRedis, TypeAMQP, TypeFirehose, TypeNoop}
