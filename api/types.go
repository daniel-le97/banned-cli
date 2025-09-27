package api

import "time"

// Channel represents a banned.video channel
type Channel struct {
	ID              string        `json:"_id"`
	Title           string        `json:"title"`
	Summary         string        `json:"summary"`
	TextInfo        string        `json:"textInfo"`
	FeaturedVideo   *Video        `json:"featuredVideo"`
	LiveStreamVideo *Video        `json:"liveStreamVideo"`
	AlwaysLive      bool          `json:"alwaysLive"`
	IsLive          bool          `json:"isLive"`
	Avatar          string        `json:"avatar"`
	CoverImage      string        `json:"coverImage"`
	Creator         *AdminUser    `json:"creator"`
	TotalVideoViews float64       `json:"totalVideoViews"`
	TotalLikes      float64       `json:"totalLikes"`
	Playlists       []Playlist    `json:"playlists"`
	PlaylistOrder   []string      `json:"playlistOrder"`
	TotalVideos     float64       `json:"totalVideos"`
	TotalPlaylists  float64       `json:"totalPlaylists"`
	Links           *ChannelLinks `json:"links"`
	LastUpload      *time.Time    `json:"lastUpload"`
	CreatedAt       *time.Time    `json:"createdAt"`
	UpdatedAt       *time.Time    `json:"updatedAt"`
	Videos          []Video       `json:"videos,omitempty"`
}

// Video represents a banned.video video
type Video struct {
	ID             string     `json:"_id"`
	Title          string     `json:"title"`
	Summary        string     `json:"summary"`
	Published      bool       `json:"published"`
	Unlisted       bool       `json:"unlisted"`
	Creator        *AdminUser `json:"creator"`
	Channel        *Channel   `json:"channel"`
	Tags           []Tag      `json:"tags"`
	VideoDuration  float64    `json:"videoDuration"`
	AudioUpload    *Upload    `json:"audioUpload"`
	PosterUpload   *Upload    `json:"posterUpload"`
	VideoUpload    *Upload    `json:"videoUpload"`
	Live           bool       `json:"live"`
	DisableComment bool       `json:"disableComment"`
	LiveStreamUrl  string     `json:"liveStreamUrl"`
	TimeWatched    float64    `json:"timeWatched"`
	IsInWatchLater bool       `json:"isInWatchLater"`
	LargeImage     string     `json:"largeImage"`
	PlayCount      float64    `json:"playCount"`
	LikeCount      float64    `json:"likeCount"`
	AngerCount     float64    `json:"angerCount"`
	ShowInfo       *ShowInfo  `json:"showInfo"`
	CreatedAt      *time.Time `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
	PublishedAt    *time.Time `json:"publishedAt"`
	LastCommentAt  *time.Time `json:"lastCommentAt"`
	DirectURL      string     `json:"directUrl"`
	EmbedURL       string     `json:"embedUrl"`
}

// AdminUser represents an admin user
type AdminUser struct {
	ID          string     `json:"_id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// Tag represents a video tag
type Tag struct {
	ID        string     `json:"_id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// Upload represents an uploaded file
type Upload struct {
	ID               string     `json:"_id"`
	Name             string     `json:"name"`
	Mimetype         string     `json:"mimetype"`
	Encoding         string     `json:"encoding"`
	OriginalFilename string     `json:"originalFilename"`
	Size             int64      `json:"size"`
	CreatedAt        *time.Time `json:"createdAt"`
	UpdatedAt        *time.Time `json:"updatedAt"`
}

// Playlist represents a playlist
type Playlist struct {
	ID          string     `json:"_id"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	Image       string     `json:"image"`
	Channel     *Channel   `json:"channel"`
	Creator     *AdminUser `json:"creator"`
	Videos      []Video    `json:"videos"`
	TotalVideos float64    `json:"totalVideos"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// ChannelLinks represents channel social links
type ChannelLinks struct {
	Website       string `json:"website"`
	Facebook      string `json:"facebook"`
	Twitter       string `json:"twitter"`
	Gab           string `json:"gab"`
	Minds         string `json:"minds"`
	Telegram      string `json:"telegram"`
	SubscribeStar string `json:"subscribeStar"`
}

// ShowInfo represents show information
type ShowInfo struct {
	Times string `json:"times"`
	Phone string `json:"phone"`
}

// Comment represents a video comment
type Comment struct {
	ID        string     `json:"_id"`
	Name      string     `json:"name"`
	User      *User      `json:"user"`
	Video     *Video     `json:"video"`
	Parent    *Comment   `json:"parent"`
	Replies   []Comment  `json:"replies"`
	LikeCount float64    `json:"likeCount"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// User represents a regular user
type User struct {
	ID          string     `json:"_id"`
	DisplayName string     `json:"displayName"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
}

// VideoView represents a video view record
type VideoView struct {
	ID           string     `json:"_id"`
	Video        *Video     `json:"video"`
	User         *User      `json:"user"`
	Channel      *Channel   `json:"channel"`
	ViewDuration float64    `json:"viewDuration"`
	UserAgent    string     `json:"userAgent"`
	IPAddress    string     `json:"ipAddress"`
	CreatedAt    *time.Time `json:"createdAt"`
}

// Vote represents a vote (like/dislike)
type Vote struct {
	ID        string     `json:"_id"`
	User      *User      `json:"user"`
	Video     *Video     `json:"video"`
	Comment   *Comment   `json:"comment"`
	Type      string     `json:"type"` // "like", "dislike", "angry"
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// VoteCount represents vote counts for content
type VoteCount struct {
	LikeCount    float64 `json:"likeCount"`
	DislikeCount float64 `json:"dislikeCount"`
	AngerCount   float64 `json:"angerCount"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// ChannelVideosParams represents parameters for fetching channel videos
type ChannelVideosParams struct {
	ChannelID          string
	IncludeUnlisted    bool
	IncludeUnpublished bool
	IncludeLive        bool
	Offset             int
	Limit              int
}

// VideosResponse represents a response containing videos with pagination info
type VideosResponse struct {
	Videos     []Video `json:"videos"`
	TotalCount int     `json:"totalCount,omitempty"`
	HasMore    bool    `json:"hasMore,omitempty"`
}

// ChannelsResponse represents a response containing channels
type ChannelsResponse struct {
	Channels   []Channel `json:"channels"`
	TotalCount int       `json:"totalCount,omitempty"`
	HasMore    bool      `json:"hasMore,omitempty"`
}
