package clients

type SpotifyArtist struct {
	Name string
}

type SpotifyAlbum struct {
	Name string
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
