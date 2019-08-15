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
			return err
		}

		if result.(int64) == 0 {
			continue
		}

		result, err = redis.Strings(q.Client.Do("ZPOPMIN", "garfunkel.queue"))
		if err != nil {
			continue
		}

		item := QueueItem{
			UserId:  strings.Split(result.([]string)[0], "-")[0],
			Service: Service(strings.Split(result.([]string)[0], "-")[1]),
		}
		c <- item

		_, err = q.Client.Do("ZADD", "garfunkel.queue", "NX", time.Now().Unix(), result.([]string)[0])
		if err != nil {
			fmt.Printf("ERROR")
			continue
		}
	}
	return nil
}
