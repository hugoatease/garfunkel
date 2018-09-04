package queue

type Service string

const (
	Spotify Service = "spotify"
)

type QueueItem struct {
	UserId  string
	Service Service
}
