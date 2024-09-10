package bemogo

import (
	"fmt"
	"sync"
	"time"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

var (
	ErrInvalidURL           = fmt.Errorf("invalid url: cannot extract the id from the url")
	ErrCannotGetVideo       = fmt.Errorf("cannot get the video: this video is restricted or does not exist")
	ErrCannotGetVideoSource = fmt.Errorf("cannot get the video source")
)

type song struct {
	videoID,
	title,
	author,
	description,
	source,
	thumbnailURL string

	duration  time.Duration
	createdAt time.Time
}

func SongFromID(videoURL string) (song, error) {
	videoID, err := youtube.ExtractVideoID(videoURL)

	if err != nil {
		fmt.Println("invalid url: ", err)
		return song{}, ErrInvalidURL
	}

	client := youtube.Client{}

	video, err := client.GetVideo(videoID)

	if err != nil {
		fmt.Println("cannot get video: ", err)
		return song{}, ErrCannotGetVideo
	}

	formats := video.Formats.WithAudioChannels()

	url, err := client.GetStreamURL(video, &formats[0])

	if err != nil {
		fmt.Println("cannot get video source: ", err)
		return song{}, err
	}

	var thumbnailURL string

	if len(video.Thumbnails) >= 1 {
		thumbnailURL = video.Thumbnails[len(video.Thumbnails)-1].URL
	}

	newSong := song{
		videoID:      videoID,
		title:        video.Title,
		author:       video.Author,
		description:  video.Description,
		source:       url,
		thumbnailURL: thumbnailURL,
		duration:     video.Duration,
		createdAt:    time.Now(),
	}

	return newSong, nil
}

func (s song) URL() string {
	return fmt.Sprintf("https://youtu.be/%s", s.videoID)
}

func (s song) Embed() *dgo.MessageEmbed {
	return &dgo.MessageEmbed{
		URL:         s.URL(),
		Title:       s.title,
		Description: fmt.Sprintf("By %s", s.author),
		Thumbnail:   &dgo.MessageEmbedThumbnail{URL: s.thumbnailURL},
		Footer:      &dgo.MessageEmbedFooter{Text: s.duration.String()},
	}
}

type playlist struct {
	bucle int

	isPlaying bool

	currentSong *song

	songs []song

	createdAt time.Time

	mu   sync.Mutex
	stop chan bool
}

func NewPlaylist() *playlist {
	songs := make([]song, 0)

	stop := make(chan bool)

	return &playlist{
		songs:     songs,
		stop:      stop,
		createdAt: time.Now(),
	}
}

func (pl *playlist) SongsEmbed() []*dgo.MessageEmbed {
	embeds := make([]*dgo.MessageEmbed, len(pl.songs))

	for i, song := range pl.songs {
		embeds[i] = song.Embed()
	}

	return embeds
}

func (pl *playlist) CurrentSong() *song {
	return pl.currentSong
}

func (pl *playlist) GetSong(i int64) *song {
	if i > int64(len(pl.songs)-1) {
		return nil
	}
	return &pl.songs[i]
}

func (pl *playlist) NextSong() *song {
	if len(pl.songs) == 0 {
		return nil
	}

	if len(pl.songs) == 1 {
		return pl.currentSong
	}

	index := 0

	for i, song := range pl.songs {
		if pl.currentSong == nil {
			break
		}

		if song.videoID == pl.currentSong.videoID {
			index = i
		}
	}

	if index == len(pl.songs)-1 {
		index = -1
	}

	index += 1

	return &pl.songs[index]
}

func (pl *playlist) PushSong(s song) {
	pl.mu.Lock()
	pl.songs = append(pl.songs, s)
	pl.mu.Unlock()
}

func (pl *playlist) PopSong(i int) {
	if len(pl.songs) == 0 {
		return
	}

	pl.mu.Lock()

	if i == -1 {
		i = len(pl.songs) - 1
	}

	pl.songs = append(pl.songs[:i], pl.songs[i+1:]...)

	pl.mu.Unlock()
}

func (pl *playlist) TogglePause() bool {
	pl.mu.Lock()
	pl.isPlaying = !pl.isPlaying
	pl.mu.Unlock()

	return pl.isPlaying
}

func (pl *playlist) JumpPrevious(n int) {}

func (pl *playlist) JumpNext(n int) {}

func (pl *playlist) SetBucle(n int) {
	pl.mu.Lock()
	pl.bucle = n
	pl.mu.Unlock()
}
