package movie

import (
	"time"
	"vger/player/shared"
)

var timerEndSeek *time.Timer

func (m *Movie) SeekOffset(offset time.Duration) {
	m.seeking.SendSeekOffset(offset)

	if timerEndSeek == nil || !timerEndSeek.Reset(200*time.Millisecond) {
		if timerEndSeek == nil {
			timerEndSeek = time.NewTimer(200 * time.Millisecond)
		}
		go func() {
			select {
			case <-timerEndSeek.C:
				m.seeking.SendEndSeekOffset()
			case <-m.quit:
				timerEndSeek.Stop()
			}
		}()
	}
}

func (m *Movie) OnSeekStarted() time.Duration {
	t := m.c.GetTime()
	m.SeekHold()

	return t / time.Second * time.Second
}

func (m *Movie) OnSeek(t time.Duration, img []byte) {
	m.w.Draw(img)
	m.SeekPlayingSubs(-1) //Hide subs while dragging
}
func (m *Movie) OnSeekPaused(t time.Duration) {
	m.SeekPlayingSubs(t)
	m.showProgressInner(t)
}
func (m *Movie) OnSeekEnded(t time.Duration) {
	if m.waitBuffer(1024 * 1024) {
		return
	}

	m.SeekUnhold(t)

	m.p.SetLastPos(t)
	shared.SavePlayingAsync(m.p)
}
