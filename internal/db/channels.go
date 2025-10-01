package db

import (
	"database/sql"
	"fmt"
)

// StoreChannel stores a channel in the database
func StoreChannel(channel Channel) error {
	db, err := GetDB()
	if err != nil {
		return err
	}

	showTimes := ""
	showPhone := ""
	if channel.ShowInfo != nil {
		showTimes = channel.ShowInfo.Times
		showPhone = channel.ShowInfo.Phone
	}

	website, facebook, twitter, gab, minds, telegram, subscribeStar := "", "", "", "", "", "", ""
	if channel.Links != nil {
		website = channel.Links.Website
		facebook = channel.Links.Facebook
		twitter = channel.Links.Twitter
		gab = channel.Links.Gab
		minds = channel.Links.Minds
		telegram = channel.Links.Telegram
		subscribeStar = channel.Links.SubscribeStar
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO channels (
			id, title, summary, text_info, avatar, cover_image, is_live,
			total_videos, total_video_views, total_likes,
			show_times, show_phone, website, facebook, twitter, gab, minds,
			telegram, subscribe_star, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, channel.ID, channel.Title, channel.Summary, channel.TextInfo, channel.Avatar,
		channel.CoverImage, channel.IsLive, int(channel.TotalVideos), 0, 0,
		showTimes, showPhone, website, facebook, twitter, gab, minds, telegram, subscribeStar)

	if err != nil {
		return fmt.Errorf("failed to store channel '%s': %w", channel.ID, err)
	}

	LogDebug("Stored channel: %s (%s)", channel.ID, channel.Title)
	return nil
}

// GetAllChannels retrieves all channels from the database
func GetAllChannels() ([]Channel, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, title, summary, text_info, avatar, cover_image, is_live,
			   COALESCE(total_videos, 0), COALESCE(total_video_views, 0), COALESCE(total_likes, 0),
			   show_times, show_phone, website, facebook, twitter, gab, minds,
			   telegram, subscribe_star
		FROM channels 
		ORDER BY title
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var channel Channel
		var totalVideos, totalVideoViews, totalLikes int
		var showTimes, showPhone, website, facebook, twitter, gab, minds, telegram, subscribeStar sql.NullString

		err := rows.Scan(
			&channel.ID, &channel.Title, &channel.Summary, &channel.TextInfo,
			&channel.Avatar, &channel.CoverImage, &channel.IsLive,
			&totalVideos, &totalVideoViews, &totalLikes,
			&showTimes, &showPhone, &website, &facebook, &twitter, &gab, &minds,
			&telegram, &subscribeStar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}

		// Assign the video counts
		channel.TotalVideos = float64(totalVideos)
		channel.TotalVideoViews = float64(totalVideoViews)
		channel.TotalLikes = float64(totalLikes)

		// Populate nested structs if data exists
		if showTimes.Valid || showPhone.Valid {
			channel.ShowInfo = &ShowInfo{
				Times: showTimes.String,
				Phone: showPhone.String,
			}
		}

		if website.Valid || facebook.Valid || twitter.Valid || gab.Valid || minds.Valid || telegram.Valid || subscribeStar.Valid {
			channel.Links = &Links{
				Website:       website.String,
				Facebook:      facebook.String,
				Twitter:       twitter.String,
				Gab:           gab.String,
				Minds:         minds.String,
				Telegram:      telegram.String,
				SubscribeStar: subscribeStar.String,
			}
		}

		channels = append(channels, channel)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading channel rows: %w", err)
	}

	return channels, nil
}
