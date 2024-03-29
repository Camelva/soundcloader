package soundcloader

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Song struct {
	client *Client

	ID           int
	Permalink    string
	PermalinkURL string
	Streams      []Stream
	Title        string
	Author       string
	Duration     time.Duration
	PublishDate  time.Time
	Thumbnail    string

	streamCounter int
}

func (s *Song) parseSongInfo(meta *metadataV2) {
	if s.client.Debug {
		log.Printf("meta: %#v", meta)
	}
	s.ID = meta.ID
	s.Permalink = meta.Permalink
	s.PermalinkURL = meta.PermalinkURL
	s.Title = meta.Title
	s.Author = meta.User.Username

	duration, _ := time.ParseDuration(fmt.Sprintf("%dms", meta.Duration))
	s.Duration = duration.Round(time.Second)

	s.Thumbnail = getBestThumbnail(meta.ArtworkURL)

	s.PublishDate = time.Date(meta.CreatedAt.Year(), meta.CreatedAt.Month(), meta.CreatedAt.Day(), 0, 0, 0, 0, time.UTC)

	s.getStreams(meta)
}

func (s *Song) getStreams(meta *metadataV2) {
	l := len(meta.Media.Transcodings)
	baseStreams := make([]Stream, 0, l)

	for _, tcoding := range meta.Media.Transcodings {
		ext := strings.SplitN(tcoding.Preset, "_", 2)[0]
		sResult := Stream{
			parent:    s,
			Format:    tcoding.Format.Protocol,
			Extension: ext,
			URL:       tcoding.URL,
		}
		sResult.Description = sResult.Extension + "-" + sResult.Format
		baseStreams = append(baseStreams, sResult)
	}

	if meta.Downloadable {
		originalStream := Stream{
			parent:      s,
			Format:      "original",
			Extension:   "",
			URL:         meta.DownloadURL,
			Description: "original",
		}
		baseStreams = append(baseStreams, originalStream)
	}
	s.Streams = sortStreams(baseStreams)
}

func (s *Song) increaseCounter() {
	if s.streamCounter >= len(s.Streams) {
		// reset counter
		s.streamCounter = 0
		return
	}
	s.streamCounter++
}

func (s *Song) GetNext() (filename string, err error) {
	if s.client.Debug {
		log.Printf("Start getting [%d] %s", s.ID, s.Permalink)
		log.Printf("Streams: %#v", s.Streams)
	}
	filename, err = s.Get(s.streamCounter)
	s.increaseCounter()

	if err == EmptyStream {
		if s.client.Debug {
			log.Printf("stream #%d empty", s.streamCounter-1)
		}
		for i := 1; i < len(s.Streams); i++ {
			if s.client.Debug {
				log.Printf("trying next stream #%d", s.streamCounter)
			}
			filename, err = s.Get(s.streamCounter)
			s.increaseCounter()

			if err == EmptyStream {
				if s.client.Debug {
					log.Printf("stream #%d empty", s.streamCounter-1)
				}
				continue
			}
			break
		}
	}
	return
}

func (s *Song) Get(i int) (filename string, err error) {
	if i >= len(s.Streams) {
		return s.Streams[len(s.Streams)-1].Get()
	} else if i < 0 {
		return s.Streams[0].Get()
	}

	if (s.Streams[i] == Stream{}) {
		return "", EmptyStream
	}

	filename, err = s.Streams[i].Get()
	if err != nil {
		return
	}

	meta := s.createMetadata()
	err = s.client.ffmpegUpdateTags(filename, meta)
	if err != nil {
		return
	}
	if s.Thumbnail != "" {
		// if we couldn't add thumbnail - its not critical, dont return this error
		e := s.client.ffmpegAddThumbnail(filename, s.Thumbnail)
		if e != nil {
			s.client.Logger.Printf("can't get song's thumbnail: %s", e)
		}
	}
	return
}

func (s *Song) GetOriginal() (filename string, err error) {
	latestEl := len(s.Streams) - 1
	if s.Streams[latestEl].Description != "original" {
		return "", NoOriginalStream
	}
	return s.Get(latestEl)
}

func (s *Song) GetOpus() (filename string, err error) {
	return s.Get(2)
}

func (s *Song) createMetadata() []string {
	metaMap := map[string]string{
		"title": s.Title,
		"album": s.Title,
		//"genre":      info.Genre,
		"artist":       s.Author,
		"album_artist": s.Author,
		"track":        strconv.Itoa(1),
		"date":         strconv.Itoa(s.PublishDate.Year()),
	}
	var metadata = make([]string, 0, len(metaMap))
	for key, value := range metaMap {
		line := fmt.Sprintf("%s=%s", key, value)
		metadata = append(metadata, line)
	}
	return metadata
}

func getBestThumbnail(u string) string {
	if u == "" {
		return ""
	}

	var bestSize = "t500x500"
	parts := strings.Split(u, "-")
	lastPartSplit := strings.Split(parts[len(parts)-1], ".")
	lastPartSplit[0] = bestSize
	parts[len(parts)-1] = strings.Join(lastPartSplit, ".")
	return strings.Join(parts, "-")
}
