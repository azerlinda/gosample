package visitor

import (
	"log"
	"time"

	redigo "github.com/garyburd/redigo/redis"
)

var redisPool *redigo.Pool

// InitRedis set redis pool
func InitRedis() {
	redisPool = NewRedis("devel-redis.tkpd:6379")
}

// SetRedis for Redis
func SetRedis(key string, value interface{}) error {
	con := redisPool.Get()
	defer con.Close()

	_, err := con.Do("SETEX", key, value, "10")
	return err
}

// GetRedis for Redis
func GetRedis(key string) (int, error) {
	con := redisPool.Get()
	defer con.Close()

	return redigo.Int(con.Do("GET", key))
}

// IncrRedis for Redis
func IncrRedis(key string) (int, error) {
	con := redisPool.Get()
	defer con.Close()

	_, err := con.Do("INCR", key)
	if err != nil {
		log.Println(err)
	}
	return redigo.Int(con.Do("GET", key))
}

// NewRedis for create redigo
func NewRedis(address string) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     1,
		IdleTimeout: 10 * time.Second,
		Dial:        func() (redigo.Conn, error) { return redigo.Dial("tcp", address) },
	}
}

// GetVisitor return visitor count from redis
func GetVisitor() (visitor int) {
	redisKey := "frans_counter"

	if val, err := GetRedis(redisKey); err != nil {
		log.Print(err)
	} else {
		log.Println("Cached")
		if _, err := IncrRedis(redisKey); err != nil {
			log.Print(err)
		}
		return val
	}

	if err := SetRedis(redisKey, 0); err != nil {
		log.Print(err)
	}
	return 0
}
