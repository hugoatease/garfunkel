package queue

/*func TestPolling(t *testing.T) {
	conn, _ := redis.DialURL("redis://")
	queue := NewQueue(conn)
	ch := make(chan QueueItem)

	go queue.Poll(ch)

	for ok := true; ok; ok = true {
		item := <-ch
		fmt.Printf("%+v\n", item)
		fmt.Printf("%+v\n", ch)
	}
}*/
