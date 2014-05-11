package movie

import (
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
	"task"
	"time"
)

type Movie struct {
	ctx AVFormatContext
	v   *Video
	a   *Audio
	s   *Subtitle
	s2  *Subtitle
	c   *Clock
	w   *Window
	p   *Playing

	chSeekPause chan time.Duration

	quit        chan bool
	finishClose chan bool

	subs []*Sub

	audioStreams []AVStream

	size int64
}

func NewMovie() *Movie {
	m := &Movie{}
	m.quit = make(chan bool)
	return m
}

func updateSubscribeDuration(movie string, duration time.Duration) {
	if t, _ := task.GetTask(movie); t != nil {
		println("get subscribe:", t.Subscribe)
		if subscr := subscribe.GetSubscribe(t.Subscribe); subscr != nil && subscr.Duration == 0 {
			subscribe.UpdateDuration(t.Subscribe, duration)
		}
	}
}

func (m *Movie) Open(w *Window, file string) {
	println("open ", file)

	var ctx AVFormatContext
	var filename string

	if strings.HasPrefix(file, "http://") {
		ctx, filename = m.openHttp(file)
		if ctx.IsNil() {
			log.Fatal("open failed: ", file)
			return
		}
		ctx.FindStreamInfo()
	} else {
		ctx = NewAVFormatContext()
		ctx.OpenInput(file)
		if ctx.IsNil() {
			log.Fatal("open failed:", file)
			return
		}

		filename = filepath.Base(file)

		ctx.FindStreamInfo()
		ctx.DumpFormat()

	}

	m.chSeekPause = make(chan time.Duration)

	m.ctx = ctx

	var duration time.Duration
	if ctx.Duration() != AV_NOPTS_VALUE {
		duration = time.Duration(float64(ctx.Duration()) / AV_TIME_BASE * float64(time.Second))
	} else {
		// duration = 2 * time.Hour
		log.Fatal("Can't get video duration.")
	}

	log.Print("video duration:", duration.String())
	m.c = NewClock(duration)

	m.setupVideo()
	m.w = w

	m.p = CreateOrGetPlaying(filename)

	var start time.Duration
	if m.p.LastPos > time.Second {
		start, _, _ = m.v.Seek(m.p.LastPos)
	}

	m.p.LastPos = start
	m.p.Duration = duration

	go updateSubscribeDuration(m.p.Movie, m.p.Duration)

	go func() {
		subs := GetSubtitlesMap(filename)
		log.Printf("%v", subs)
		if len(subs) == 0 {
			m.SearchDownloadSubtitle()
		} else {
			println("setupSubtitles")
			m.setupSubtitles(subs)

			if m.s != nil {
				m.s.Seek(m.c.GetTime())
			}
			if m.s2 != nil {
				m.s2.Seek(m.c.GetTime())
			}
		}
	}()

	w.InitEvents()
	w.SetTitle(filename)
	w.SetSize(m.v.Width, m.v.Height)
	m.v.SetRender(m.w)

	m.setupAudio()

	m.uievents()

	m.c.SetTime(start)

	go m.showProgress(filename)

	w.HideCursor()
}

func (m *Movie) SavePlaying() {
	SavePlaying(m.p)
}

func (m *Movie) Close() {
	m.w.FlushImageBuffer()
	m.w.RefreshContent(nil)
	m.w.ShowStartupView()

	m.finishClose = make(chan bool)
	close(m.quit)
	// time.Sleep(100 * time.Millisecond)

	m.w.ClearEvents()

	if m.s != nil {
		m.s.Stop()
		m.s = nil
	}

	if m.s2 != nil {
		m.s2.Stop()
		m.s2 = nil
	}

	<-m.finishClose
}
func (m *Movie) PlayAsync() {
	go m.v.Play()
	go m.decode(m.p.Movie)
}
func (m *Movie) Resume() {
	m.c.Resume()
}
func (m *Movie) Pause() {
	m.c.Pause()
}

func tabs(t time.Duration) time.Duration {
	if t < 0 {
		t = -t
	}
	return t
}

func (m *Movie) setupVideo() {
	ctx := m.ctx
	videoStream := ctx.VideoStream()
	if !videoStream.IsNil() {
		var err error
		m.v, err = NewVideo(ctx, videoStream, m.c)
		if err != nil {
			log.Fatal(err)
			return
		}

	} else {
		log.Fatal("No video stream find.")
	}
}

func (m *Movie) SendPacket(index int, ch chan *AVPacket, packet AVPacket) bool {
	if index == packet.StreamIndex() {
		pkt := packet
		pkt.Dup()

		select {
		case ch <- &pkt:
			return true
		case <-m.quit:
			return false
		}
	}
	return false
}
func (m *Movie) showProgress(name string) {
	m.p.LastPos = m.c.GetTime()

	p := m.c.CalcPlayProgress(m.c.GetPercent())

	done := make(chan struct{})
	go func() {
		t, err := task.GetTask(name)

		if err == nil {
			if t.Status == "Finished" {
				p.Percent2 = 1
			} else {
				p.Percent2 = float64(t.BufferedPosition) / float64(t.Size)
			}
		} else {
			log.Print(err)
		}
		close(done)
	}()

	select {
	case <-done:
		break
	case <-time.After(100 * time.Millisecond):
		break
	}

	m.w.SendShowProgress(p)
}

func (m *Movie) decode(name string) {
	defer func() {
		if m.a != nil {
			m.a.Close()
		}
		if m.v != nil {
			m.v.Close()
		}
		m.c.Reset()
		m.ctx.CloseInput()

		if m.finishClose != nil {
			close(m.finishClose)
		}
	}()

	packet := AVPacket{}
	ctx := m.ctx
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			if m.c.WaitUtilRunning(m.quit) {
				return
			}

			select {
			case <-ticker.C:
				m.showProgress(name)
			case <-m.quit:
				return
			}
		}
	}()

	bufferring := false
	for {
		resCode := ctx.ReadFrame(&packet)
		if resCode >= 0 {
			if bufferring {
				bufferring = false
				m.c.Resume()
			}
			if m.v.StreamIndex == packet.StreamIndex() {
				if frameFinished, pts, img := m.v.DecodeAndScale(&packet); frameFinished {
					//make sure seek operations not happens before one frame finish decode
					//if not, segment fault & crash
					select {
					case m.v.ChanDecoded <- &VideoFrame{pts, img}:
						break
					case t := <-m.chSeekPause:
						if t != -1 {
							break
						}
						for {
							t := <-m.chSeekPause
							if t >= 0 {
								m.c.SetTime(t)
								break
							}
						}
						break
					case <-m.quit:
						packet.Free()
						return
					}

					t := m.c.GetTime()
					if m.s != nil {
						m.s.Seek(t)
					}
					if m.s2 != nil {
						m.s2.Seek(t)
					}
				}
				packet.Free()
				continue
			}

			if m.a != nil {
				if m.SendPacket(m.a.StreamIndex(), m.a.PacketChan, packet) {
					continue
				}
			}

			packet.Free()
		} else {
			bufferring = true
			m.c.Pause()

			m.a.FlushBuffer()
			m.v.FlushBuffer()

			t, _, err := m.v.Seek(m.c.GetTime())
			if err == nil {
				println("seek success:", t.String())
				m.c.SetTime(t)
				continue
			} else {
				log.Print("seek error:", err)
			}

			// println("seek to unfinished:", m.c.GetTime().String())
			log.Print("get frame error:", resCode)

			select {
			case t := <-m.chSeekPause:
				println("seek to unfinished2")
				if t != -1 {
					continue
				}
				for {
					println("seek to unfinished3")
					t := <-m.chSeekPause
					println("seek to unfinished4")
					if t >= 0 {
						m.c.SetTime(t)
						break
					}
				}
			case <-time.After(100 * time.Millisecond):
				break
			case <-m.quit:
				return
			}

		}
		// println(bufferring)
	}
}