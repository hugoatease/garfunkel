package clients

type Client interface {
	GetCurrentlyPlaying(string) (*Listen, error)
}

type Listen struct {
	ArtistName string
	AlbumName  string
	TrackName  string
	Timestamp  int64
	ImageURL   string
}

type SpotifyArtist struct {
	Name string
}

type SpotifyImage struct {
	Height int64
	Width  int64
	URL    string
}

type SpotifyAlbum struct {
	Name   string
	Images []SpotifyImage
}

type SpotifyTrack struct {
	Name    string
	Album   SpotifyAlbum
	Artists []SpotifyArtist
}

type SpotifyListen struct {
	Timestamp int64
	Item      SpotifyTrack
}

type SpotifyRefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type DeezerArtist struct {
	Name string
}

type DeezerAlbum struct {
	Title string
}

type DeezerTrack struct {
	Timestamp int64
	Artist    DeezerArtist
	Album     DeezerAlbum
	Title     string
}

type DeezerHistory struct {
	Data []DeezerTrack
}
