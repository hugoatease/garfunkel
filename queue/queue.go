package queue

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
)

type Queue struct {
	Client *redis.Client
}

func NewQueue(options redis.Options) *Queue {
	return &Queue{
		Client: redis.NewClient(&options),
	}
}

func (q Queue) Poll(c chan QueueItem) error {
	items, err := q.Client.ZRangeWithScores("garfunkel.queue", 0, 0).Result()
	if err != nil {
		return err
	}

	for _, item := range items {
		item.Score = float64(time.Now().Unix())
		_, err := q.Client.ZAddXX("grafunkel.queue", item).Result()
		if err != nil {
			return err
		}

		itemString, ok := item.Member.(string)
		if ok {
			queueItem := QueueItem{
				UserId:  strings.Split(itemString, "-")[0],
				Service: Service(strings.Split(itemString, "-")[1]),
			}
			c <- queueItem
		}
	}

	return q.Poll(c)
}
