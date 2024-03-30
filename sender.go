package main

import (
	"fmt"
	"net/http"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func getPhotoByUrl(photo Image) (*http.Response, error) {
	file, error := http.Get(photo.Url)
	if error != nil {
		return nil, error
	}
	return file, error
}

type VideoFile struct {
	response *http.Response
	Url      string
	size     int64
}

// Method checks video size and if size more than telegram limits - returns just link else returns buffer for file downloading
func getVideoByUrl(video Video) (VideoFile, error) {
	resp, err := http.Head(video.Url)
	if err != nil {
		return VideoFile{}, err
	}
	defer resp.Body.Close()
	if resp.ContentLength > 5e7 {
		return VideoFile{}, fmt.Errorf("video size limit: %v", resp.ContentLength)
	} else if resp.ContentLength > 2e7 {
		resp, err := http.Get(video.Url)
		if err != nil {
			return VideoFile{}, err
		}
		return VideoFile{response: resp, Url: video.Url, size: resp.ContentLength}, nil
	}
	return VideoFile{response: nil, Url: video.Url, size: resp.ContentLength}, nil
}

func sendImages(caption string, story *Story, b *gotgbot.Bot, ctx *ext.Context) error {

	if len(story.Images) == 0 {
		return nil
	} else if len(story.Images) == 1 {
		var content gotgbot.InputFile
		if story.Images[0].FileId == "" {
			file, error := getPhotoByUrl(story.Images[0])
			if error != nil {
				return fmt.Errorf("failed to handle pikabu Url: %w", error)
			}
			defer file.Body.Close()
			content = file.Body
		} else {
			content = story.Images[0].FileId
		}
		sent, err := b.SendPhoto(ctx.Message.Chat.Id, content, &gotgbot.SendPhotoOpts{ParseMode: "HTML", Caption: fmt.Sprintf(caption, story.Url, story.Title)})
		if err == nil {
			story.Images[0].FileId = sent.Photo[0].FileId
		}
	} else {
		mediaGroups := [][]gotgbot.InputMedia{}
		mediaGroup := make([]gotgbot.InputMedia, 0, 10)
		for _, image := range story.Images {
			var content gotgbot.InputFile
			if image.FileId == "" {
				file, error := getPhotoByUrl(image)
				if error != nil {
					return fmt.Errorf("failed to handle pikabu Url: %w", error)
				}
				defer file.Body.Close()
				content = file.Body
			} else {
				content = image.FileId
			}
			mediaGroup = append(mediaGroup, gotgbot.InputMediaPhoto{ParseMode: "HTML", Media: content, Caption: fmt.Sprintf(caption, story.Url, story.Title)})
			if len(mediaGroup) == 10 {
				mediaGroups = append(mediaGroups, mediaGroup)
				mediaGroup = []gotgbot.InputMedia{}
			}
		}
		if len(mediaGroup) > 0 {
			mediaGroups = append(mediaGroups, mediaGroup)
		}
		imageCounter := 0
		for _, mediaGroup := range mediaGroups {
			sent, err := b.SendMediaGroup(ctx.Message.Chat.Id, mediaGroup, &gotgbot.SendMediaGroupOpts{})
			if err == nil {
				for _, photo := range sent {
					if story.Images[imageCounter].FileId == "" {
						story.Images[imageCounter].FileId = photo.Photo[0].FileId
						imageCounter++
					}
				}
			} else {
				fmt.Println("Error ", err)
			}
		}
	}
	return nil
}

func sendVideos(caption string, story Story, b *gotgbot.Bot, ctx *ext.Context) error {
	if len(story.Videos) == 0 {
		return nil
	} else if len(story.Videos) == 1 {
		var file gotgbot.InputFile
		if story.Videos[0].FileId == "" {
			video, error := getVideoByUrl(story.Videos[0])
			if error != nil {
				return fmt.Errorf("failed to handle pikabu Url: %w", error)
			}

			if video.response != nil {
				file = video.response.Body
				defer video.response.Body.Close()
			} else {
				file = video.Url
			}
		} else {
			file = story.Videos[0].FileId
		}
		sent, err := b.SendVideo(ctx.Message.Chat.Id, file, &gotgbot.SendVideoOpts{ParseMode: "HTML", Caption: fmt.Sprintf(caption, story.Url, story.Title), Duration: int64(story.Videos[0].Duration)})
		if err == nil {
			story.Videos[0].FileId = sent.Video.FileId
		}
	} else {
		mediaGroups := [][]gotgbot.InputMedia{}
		mediaGroup := make([]gotgbot.InputMedia, 0, 10)
		for _, video := range story.Videos {
			var file gotgbot.InputFile
			if story.Videos[0].FileId == "" {
				videoFile, error := getVideoByUrl(video)
				if error != nil {
					return fmt.Errorf("failed to handle pikabu Url: %w", error)
				}

				if videoFile.response != nil {
					file = videoFile.response.Body
					defer videoFile.response.Body.Close()
				} else {
					file = videoFile.Url
				}
			} else {
				file = video.FileId
			}
			mediaGroup = append(mediaGroup, gotgbot.InputMediaVideo{ParseMode: "HTML", Media: file, Caption: fmt.Sprintf(caption, story.Url, story.Title),
				Duration: int64(video.Duration)})
			if len(mediaGroup) == 10 {
				mediaGroups = append(mediaGroups, mediaGroup)
				mediaGroup = []gotgbot.InputMedia{}
			}
		}
		if len(mediaGroup) > 0 {
			mediaGroups = append(mediaGroups, mediaGroup)
		}
		videoCounter := 0
		for _, mediaGroup := range mediaGroups {
			sent, err := b.SendMediaGroup(ctx.Message.Chat.Id, mediaGroup, &gotgbot.SendMediaGroupOpts{})
			if err == nil {
				for _, video := range sent {
					if story.Videos[videoCounter].FileId == "" {
						story.Videos[videoCounter].FileId = video.Video.FileId
						videoCounter++
					}
				}
			} else {
				fmt.Println("Error ", err)
			}
		}
	}
	return nil
}
