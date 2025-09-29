package cmd

// Import the db package to make database functions available
import (
	"github.com/daniel-le97/banned-cli/db"
	bolt "go.etcd.io/bbolt"
)

// Database function aliases for backward compatibility
// These functions call the db package directly

func GetDB() (*bolt.DB, error) {
	return db.GetDB()
}

func CloseDB() error {
	return db.CloseDB()
}

func StoreChannel(channel Channel) error {
	// Convert cmd.Channel to db.Channel
	dbChannel := db.Channel{
		ID:              channel.ID,
		Title:           channel.Title,
		Summary:         channel.Summary,
		TextInfo:        channel.TextInfo,
		Avatar:          channel.Avatar,
		CoverImage:      channel.CoverImage,
		IsLive:          channel.IsLive,
		TotalVideos:     channel.TotalVideos,
		TotalVideoViews: channel.TotalVideoViews,
		TotalLikes:      channel.TotalLikes,
		ShowInfo:        convertCmdShowInfoToDB(channel.ShowInfo),
		Links:           convertCmdLinksToDB(channel.Links),
	}
	return db.StoreChannel(dbChannel)
}

func StoreVideos(channelID string, videos []Video) error {
	// Convert []cmd.Video to []db.Video
	dbVideos := make([]db.Video, len(videos))
	for i, v := range videos {
		dbVideos[i] = db.Video{
			ID:            v.ID,
			Title:         v.Title,
			Summary:       v.Summary,
			LargeImage:    v.LargeImage,
			VideoDuration: v.VideoDuration,
			CreatedAt:     v.CreatedAt,
			DirectURL:     v.DirectURL,
			PlayCount:     v.PlayCount,
			LikeCount:     v.LikeCount,
			AngerCount:    v.AngerCount,
			EmbedURL:      v.EmbedURL,
			Published:     v.Published,
			FileSize:      v.FileSize,
		}
	}
	return db.StoreVideos(channelID, dbVideos)
}

func GetChannelVideos(channelID string, limit, offset int) ([]Video, error) {
	dbVideos, err := db.GetChannelVideos(channelID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert []db.Video to []cmd.Video
	videos := make([]Video, len(dbVideos))
	for i, v := range dbVideos {
		videos[i] = Video{
			ID:            v.ID,
			Title:         v.Title,
			Summary:       v.Summary,
			LargeImage:    v.LargeImage,
			VideoDuration: v.VideoDuration,
			CreatedAt:     v.CreatedAt,
			DirectURL:     v.DirectURL,
			PlayCount:     v.PlayCount,
			LikeCount:     v.LikeCount,
			AngerCount:    v.AngerCount,
			EmbedURL:      v.EmbedURL,
			Published:     v.Published,
			FileSize:      v.FileSize,
		}
	}
	return videos, nil
}

func GetAllChannels() ([]Channel, error) {
	dbChannels, err := db.GetAllChannels()
	if err != nil {
		return nil, err
	}

	// Convert []db.Channel to []cmd.Channel
	channels := make([]Channel, len(dbChannels))
	for i, c := range dbChannels {
		channels[i] = Channel{
			ID:              c.ID,
			Title:           c.Title,
			Summary:         c.Summary,
			TextInfo:        c.TextInfo,
			Avatar:          c.Avatar,
			CoverImage:      c.CoverImage,
			IsLive:          c.IsLive,
			TotalVideos:     c.TotalVideos,
			TotalVideoViews: c.TotalVideoViews,
			TotalLikes:      c.TotalLikes,
			ShowInfo:        convertDBShowInfoToCmd(c.ShowInfo),
			Links:           convertDBLinksToCmd(c.Links),
		}
	}
	return channels, nil
}

func UpdateVideoFileSize(videoID string, fileSize int64) error {
	return db.UpdateVideoFileSize(videoID, fileSize)
}

func GetSetting(key string) (string, error) {
	return db.GetSetting(key)
}

func SetSetting(key, value string) error {
	return db.SetSetting(key, value)
}

func getDatabasePath() (string, error) {
	return db.GetDatabasePath()
}

func GetBucketStats() (map[string]int, error) {
	return db.GetBucketStats()
}

func CountChannels() (int, error) {
	return db.CountChannels()
}

func CountVideos() (int, error) {
	return db.CountVideos()
}

func CountVideosForChannel(channelID string) (int, error) {
	return db.CountVideosForChannel(channelID)
}

func GetVideosWithoutFileSize(channelID string) ([]Video, error) {
	dbVideos, err := db.GetVideosWithoutFileSize(channelID)
	if err != nil {
		return nil, err
	}

	// Convert db.Video to cmd.Video
	var videos []Video
	for _, dbVideo := range dbVideos {
		video := Video{
			ID:            dbVideo.ID,
			Title:         dbVideo.Title,
			Summary:       dbVideo.Summary,
			LargeImage:    dbVideo.LargeImage,
			VideoDuration: dbVideo.VideoDuration,
			CreatedAt:     dbVideo.CreatedAt,
			DirectURL:     dbVideo.DirectURL,
			PlayCount:     dbVideo.PlayCount,
			LikeCount:     dbVideo.LikeCount,
			AngerCount:    dbVideo.AngerCount,
			EmbedURL:      dbVideo.EmbedURL,
			Published:     dbVideo.Published,
			FileSize:      dbVideo.FileSize,
		}
		videos = append(videos, video)
	}

	return videos, nil
}

func UpdateVideoFileSizeByURL(directURL string, fileSize int64) error {
	return db.UpdateVideoFileSizeByURL(directURL, fileSize)
}

func UpdateDownloadTorrentInfo(filePath, torrentPath string) error {
	return db.UpdateDownloadTorrentInfo(filePath, torrentPath)
}

// Conversion helper functions
func convertCmdShowInfoToDB(info *ShowInfo) *db.ShowInfo {
	if info == nil {
		return nil
	}
	return &db.ShowInfo{
		Times: info.Times,
		Phone: info.Phone,
	}
}

func convertCmdLinksToDB(links *Links) *db.Links {
	if links == nil {
		return nil
	}
	return &db.Links{
		Website:       links.Website,
		Facebook:      links.Facebook,
		Twitter:       links.Twitter,
		Gab:           links.Gab,
		Minds:         links.Minds,
		Telegram:      links.Telegram,
		SubscribeStar: links.SubscribeStar,
	}
}

func convertDBShowInfoToCmd(info *db.ShowInfo) *ShowInfo {
	if info == nil {
		return nil
	}
	return &ShowInfo{
		Times: info.Times,
		Phone: info.Phone,
	}
}

func convertDBLinksToCmd(links *db.Links) *Links {
	if links == nil {
		return nil
	}
	return &Links{
		Website:       links.Website,
		Facebook:      links.Facebook,
		Twitter:       links.Twitter,
		Gab:           links.Gab,
		Minds:         links.Minds,
		Telegram:      links.Telegram,
		SubscribeStar: links.SubscribeStar,
	}
}
