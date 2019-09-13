package queue

import (
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Queue struct {
	Client redis.Conn
	Ticker *time.Ticker
}

var (
	queueScript *redis.Script = redis.NewScript(1, "local r = redis.call('ZPOPMIN', KEYS[1])[1]; redis.call('ZADD', KEYS[1], 'NX', ARGV[1], r); return r;")
)

func NewQueue(conn redis.Conn, delay time.Duration) *Queue {
	ticker := time.NewTicker(delay)
	return &Queue{
		Client: conn,
		Ticker: ticker,
	}
}

func (q Queue) Poll(c chan QueueItem) error {
	for range q.Ticker.C {
		result, err := q.Client.Do("ZCOUNT", "garfunkel.queue", 0, time.Now().Add(-10*time.Second).Unix())
		if err != nil {
			fmt.Printf("%s", err)
			return err
		}

		if result.(int64) == 0 {
			continue
		}

		result, err = redis.String(queueScript.Do(q.Client, "garfunkel.queue", time.Now().Unix()))
		if err != nil {
			fmt.Printf("%s", err)
			fmt.Printf("Error")
			continue
		}

		item := QueueItem{
			UserId:  strings.Split(result.(string), "-")[0],
			Service: Service(strings.Split(result.(string), "-")[1]),
		}
		c <- item
	}
	return nil
}
