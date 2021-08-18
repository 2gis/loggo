package redisclient

import (
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/2gis/loggo/configuration"
	"github.com/2gis/loggo/transport"
)

// RedisClient contains underlying redigo pool and follows transport interface
type RedisClient struct {
	pool *redis.Pool
	key  string
}

// NewRedisClient is a constructor for RedisClient
func NewRedisClient(config configuration.RedisTransportConfig) *RedisClient {
	dialFunction := func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", config.URL)

		if err != nil {
			return nil, err
		}

		return c, nil
	}
	testFunction := func(c redis.Conn, t time.Time) error {
		if time.Since(t) < time.Minute {
			return nil
		}

		_, err := c.Do("PING")
		return err
	}
	return &RedisClient{
		key: config.Key,
		pool: &redis.Pool{
			Dial:         dialFunction,
			TestOnBorrow: testFunction,
			MaxIdle:      transport.RedisMaxIdleConnections,
		},
	}
}

// DeliverMessages tries to send slice of messages to Redis, acquiring connection from pool
func (client *RedisClient) DeliverMessages(messages []string) error {
	sendList := make([]interface{}, 0, 1001)
	connection := client.pool.Get()
	sendList = append(sendList, client.key)

	for _, value := range messages {
		sendList = append(sendList, value)
	}

	_, err := connection.Do("RPUSH", sendList...)
	connection.Close()
	return err
}

// ReceiveMessage returns message from list, only for test purposes for now
func (client *RedisClient) ReceiveMessage() ([]byte, error) {
	connection := client.pool.Get()
	defer connection.Close()
	message, err := redis.Bytes(connection.Do("LPOP", client.key))

	if err != nil {
		return nil, err
	}

	return message, nil
}

// Close releases pool resources
func (client *RedisClient) Close() error {
	return client.pool.Close()
}
