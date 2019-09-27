package queue

type Service string

const (
	Spotify Service = "spotify"
	Deezer  Service = "deezer"
)

type QueueItem struct {
	UserId  string
	Service Service
}
