package movie

import (
	"block"
	"download"
	"fmt"
	"log"
	"path/filepath"
	. "player/audio"
	. "player/clock"
	. "player/gui"
	. "player/libav"
	. "player/shared"
	. "player/subtitle"
	. "player/video"
	"strings"
	"subscribe"
	"sync"
	"task"
	"time"
)

type Movie struct {
	ctx AVFormatContext
	v   *Video
	a   *Audio
	c   *Clock
	w   *Window
	p   *Playing
	movieSubs

	quit        chan struct{}
	finishClose chan bool

	subs []*Subtitle

	audioStreams []AVStream

	httpBuffer *buffer

	// chSeekEnd      chan time.Duration
	chSeekProgress chan *seekArg
	chPause        chan chan time.Duration
	chProgress     chan time.Duration
	chSpeed        chan float64
	streaming      *download.Streaming
}

type movieSubs struct {
	sync.Mutex

	s1 *Subtitle
	s2 *Subtitle
}

func (m *movieSubs) getPlayingSubs() (*Subtitle, *Subtitle) {
	m.Lock()
	defer m.Unlock()

	return m.s1, m.s2
}
func (m *movieSubs) setPlayingSubs(s1, s2 *Subtitle) {
	m.Lock()
	defer m.Unlock()

	m.s1, m.s2 = s1, s2
}
func (m *movieSubs) seekPlayingSubs(t time.Duration, refresh bool) {
	s1, s2 := m.getPlayingSubs()
	if s1 != nil {
		s1.Seek(t, refresh)
	}
	if s2 != nil {
		s2.Seek(t, refresh)
	}
}
func (m *movieSubs) stopPlayingSubs() {
	s1, s2 := m.getPlayingSubs()
	m.setPlayingSubs(nil, nil)

	if s1 != nil {
		s1.Stop()
	}
	if s2 != nil {
		s2.Stop()
	}
}

type seekArg struct {
	t     time.Duration
	isEnd bool
}

func NewMovie() *Movie {
	log.Print("New movie")

	m := &Movie{}
	m.quit = make(chan struct{})
	m.chProgress = make(chan time.Duration)
	m.finishClose = make(chan bool)
	return m
}

func updateSubscribeDuration(movie string, duration time.Duration) {
	if t, _ := task.GetTask(movie); t != nil {
		if subscr := subscribe.GetSubscribe(t.Subscribe); subscr != nil && subscr.Duration == 0 {
			err := subscribe.UpdateDuration(t.Subscribe, duration)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func checkDownloadSubtitle(m *Movie, file string, filename string) {
	if strings.Contains(file, "googlevideo.com/videoplayback") {
		return
	}

	subs := GetSubtitlesMap(filename)
	log.Printf("%v", subs)
	if len(subs) == 0 {
		m.SearchDownloadSubtitle()
	} else {
		log.Print("setupSubtitles")
		m.setupSubtitles(subs)

		m.seekPlayingSubs(m.c.GetTime(), false)
	}
}

func setBufferCapacity(buf *buffer, duration time.Duration) {
	var capacity int64
	if duration < 10*time.Minute {
		capacity = buf.size
	} else {
		//overflow, divide before cross
		capacity = int64(float64(buf.size) / float64(duration) * 5 * float64(time.Minute))
		log.Print(capacity)
	}

	if capacity > 100*block.MB {
		capacity = 100 * block.MB
	}
	if capacity < 10*block.MB {
		capacity = 10 * block.MB
	}
	buf.SetCapacity(capacity)
}

func (m *Movie) setupContext(file string) (filename string, duration time.Duration, err error) {
	var ctx AVFormatContext

	if strings.HasPrefix(file, "http://") ||
		strings.HasPrefix(file, "https://") {

		ctx, filename, err = m.openHttp(file)
		if err != nil {
			err = fmt.Errorf("open failed: %s", file)
			return
		}
	} else {
		filename = filepath.Base(file)

		ctx = NewAVFormatContext()
		if err = ctx.OpenInput(file); err != nil {
			return
		}
	}

	if err = ctx.FindStreamInfo(); err != nil {
		return
	}
	ctx.DumpFormat()
	m.ctx = ctx

	duration = ctx.Duration()
	return
}

func (m *Movie) Open(w *Window, file string) (err error) {
	defer func() {
		if err != nil {
			close(m.finishClose)
		}
	}()

	log.Print("open ", file)

	m.w = w
	m.uievents()

	filename, duration, err := m.setupContext(file)
	if err != nil {
		return
	}

	if m.httpBuffer != nil {
		setBufferCapacity(m.httpBuffer, duration)
	}

	m.c = NewClock(duration)

	m.p = CreateOrGetPlaying(filename)
	m.p.Duration = duration

	err = m.setupVideo()
	if err != nil {
		return
	}

	err = m.setupAudio()
	if err != nil {
		return
	}

	go updateSubscribeDuration(m.p.Movie, m.p.Duration)
	go checkDownloadSubtitle(m, file, filename)

	w.SendSetTitle(filename)
	w.SendSetCursor(false)

	return
}

func (m *Movie) SavePlaying() {
	if m.p != nil {
		SavePlaying(m.p)
	}
}

func (m *Movie) Close() {
	// m.w.Destory()
	m.w.ShowStartupView()

	close(m.quit)

	m.w.ClearEvents()

	m.stopPlayingSubs()

	HideSubtitleMenu()
	HideAudioMenu()

	if m.httpBuffer != nil {
		m.httpBuffer.Close()
	}

	if m.streaming != nil {
		m.streaming.Close()
	}

	<-m.finishClose

	m.w.SendDestoryRender()
}

func (m *Movie) PlayAsync() {
	log.Print("movie play async")

	go m.v.Play()
	go m.showProgressPerSecond()
	go m.decode(m.p.Movie)
}

func (m *Movie) setupVideo() error {
	log.Print("setup video")

	ctx := m.ctx
	videoStream := ctx.VideoStream()
	if !videoStream.IsNil() {
		var err error
		m.v, err = NewVideo(ctx, videoStream, m.c, m.w)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("No video stream find.")
	}

	return nil
}

func (m *Movie) ResumeClock() {
	m.c.Resume()
}

func (m *Movie) PauseClock() {
	m.c.Pause()
}
