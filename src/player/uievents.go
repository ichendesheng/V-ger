package main

import (
	"player/gui"
	. "player/libav"
	"time"
)

func (m *movie) uievents() {
	m.v.window.FuncKeyDown = append(m.v.window.FuncKeyDown, func(keycode int) {
		switch keycode {
		case gui.KEY_SPACE:
			m.c.Toggle()
			break
		case gui.KEY_LEFT:
			println("key left pressed")
			m.c.SetTime(m.SeekTo(m.c.GetSeekTime() - 10*time.Second))
			break
		case gui.KEY_RIGHT:
			m.c.SetTime(m.SeekTo(m.c.GetSeekTime() + 10*time.Second))
			break
		case gui.KEY_UP:
			m.c.SetTime(m.SeekTo(m.c.GetSeekTime() + time.Minute))
			break
		case gui.KEY_DOWN:
			m.c.SetTime(m.SeekTo(m.c.GetSeekTime() - time.Minute))
			break
		case gui.KEY_MINUS:
			println("key minus pressed")
			if m.s != nil {
				m.s.AddOffset(-1000 * time.Millisecond)
			}
			break
		case gui.KEY_EQUAL:
			println("key equal pressed")
			if m.s != nil {
				m.s.AddOffset(1000 * time.Millisecond)
			}
			break
		case gui.KEY_LEFT_BRACKET:
			println("left bracket pressed")
			if m.s != nil {
				m.s.AddOffset(-200 * time.Millisecond)
			}
			break
		case gui.KEY_RIGHT_BRACKET:
			println("right bracket pressed")
			if m.s != nil {
				m.s.AddOffset(200 * time.Millisecond)
			}
			break
		}
	})

	var lastSeekTime time.Duration
	var lastText uintptr

	m.v.window.FuncOnProgressChanged = append(m.v.window.FuncOnProgressChanged, func(typ int, percent float64) { //run in main thread, safe to operate ui elements
		switch typ {
		case 0:
			lastSeekTime = m.c.GetSeekTime()

			m.c.Pause()
			break
		case 2:
			if lastText != 0 {
				m.v.window.HideText(lastText)
				lastText = 0
			}
			t := m.c.CalcTime(percent)
			m.c.ResumeWithTime(m.SeekTo(t))
			break
		case 1:
			t := m.c.CalcTime(percent)
			flags := AVSEEK_FLAG_FRAME
			if t < lastSeekTime {
				flags |= AVSEEK_FLAG_BACKWARD
			}
			m.ctx.SeekFrame(m.v.stream, t, flags)
			lastSeekTime = t

			codec := m.v.stream.Codec()
			codec.FlushBuffer()
			m.drawCurrentFrame()

			if m.s != nil {
				if _, item := m.s.FindPos(t); item != nil {
					if lastText != 0 {
						m.v.window.HideText(lastText)
						lastText = 0
					}

					lastText = m.v.window.ShowText(item)
				} else {
					if lastText != 0 {
						m.v.window.HideText(lastText)
						lastText = 0
					}
				}
			}

			break
		}

		m.v.window.ShowProgress(m.c.CalcPlayProgress(percent))
	})
}
