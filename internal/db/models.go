package db

// Video represents a video structure
type Video struct {
	ID            string  `json:"_id"`
	Title         string  `json:"title"`
	Summary       string  `json:"summary"`
	LargeImage    string  `json:"largeImage"`
	VideoDuration float64 `json:"videoDuration"`
	CreatedAt     string  `json:"createdAt"`
	DirectURL     string  `json:"directUrl"`
	PlayCount     int     `json:"playCount,omitempty"`
	LikeCount     int     `json:"likeCount,omitempty"`
	AngerCount    int     `json:"angerCount,omitempty"`
	EmbedURL      string  `json:"embedUrl,omitempty"`
	Published     bool    `json:"published,omitempty"`
	FileSize      int64   `json:"fileSize,omitempty"`
}

// Channel represents a channel structure
type Channel struct {
	ID              string    `json:"_id"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary,omitempty"`
	TextInfo        string    `json:"textInfo,omitempty"`
	Avatar          string    `json:"avatar,omitempty"`
	CoverImage      string    `json:"coverImage,omitempty"`
	IsLive          bool      `json:"isLive,omitempty"`
	TotalVideos     float64   `json:"totalVideos,omitempty"`
	TotalVideoViews float64   `json:"totalVideoViews,omitempty"`
	TotalLikes      float64   `json:"totalLikes,omitempty"`
	Videos          []Video   `json:"videos,omitempty"`
	ShowInfo        *ShowInfo `json:"showInfo,omitempty"`
	Links           *Links    `json:"links,omitempty"`
}

type ShowInfo struct {
	Times string `json:"times"`
	Phone string `json:"phone"`
}

type Links struct {
	Website       string `json:"website"`
	Facebook      string `json:"facebook"`
	Twitter       string `json:"twitter"`
	Gab           string `json:"gab"`
	Minds         string `json:"minds"`
	Telegram      string `json:"telegram"`
	SubscribeStar string `json:"subscribeStar"`
}
