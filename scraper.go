package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gocolly/colly"
)

func parsePage(url string, pikabuID int) (Story, error) {
	var result Story
	result.Url = url
	result.PikabuID = pikabuID
	c := colly.NewCollector()
	// Find title of story
	c.OnHTML("span.story__title-link", func(e *colly.HTMLElement) {
		if result.Title == "" {
			result.Title = e.Text
		}
	})
	// Add parsing for images and video
	images, videos := []Image{}, []Video{}
	parseContent(c, &images, &videos)
	err := c.Visit(url)
	if err != nil {
		fmt.Println(err)
		return result, err
	}
	result.Images = images;
	result.Videos = videos;
	return result, nil
}

func parseContent(c *colly.Collector, images *[]Image, videos *[]Video) {
	c.OnHTML("div.page-story__story", func(e *colly.HTMLElement) {
		e.ForEach("div.story-image__content", func(_ int, h *colly.HTMLElement) {
			h.ForEach("a.image-link", func(_ int, h *colly.HTMLElement) {
				*images = append(*images, Image{Url: h.Attr("href")})
			})
		})
		e.ForEach("div.player", func(_ int, h *colly.HTMLElement) {
			if h.Attr("data-type") == "video-file" {
				video_duration, err := strconv.Atoi(h.Attr("data-duration"))
				if err != nil {
					video_duration = 0
					log.Println(err)
				}
				*videos = append(*videos, Video{Url: h.Attr("data-source") + ".mp4", Duration: video_duration})
			}
		})
	})
}
