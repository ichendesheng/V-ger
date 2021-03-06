package movie

import (
	"fmt"
	"log"
	"os"
	"time"
	"vger/download"
	"vger/player/libav"
	"vger/util"
)

func (m *Movie) waitBuffer(size int64) bool {
	if m.httpBuffer != nil {
		m.w.SendShowSpinning()
		defer m.w.SendHideSpinning(false)

		return m.httpBuffer.WaitQuit(size, m.quit)
	}

	return false
}

func (m *Movie) openHttp(file string) (libav.AVFormatContext, string, error) {
	download.NetworkTimeout = time.Duration(util.ReadIntConfig("network-timeout")) * time.Second
	download.BaseDir = util.ReadConfig("dir")

	m.chSpeed = make(chan float64)

	url, name, size, _, err := download.GetDownloadInfoN(file, nil, 10, false, m.quit)

	if err != nil {
		return libav.AVFormatContext{}, "", err
	}

	m.httpBuffer = newBuffer(size)

	buf := libav.AVObject{}
	buf.Malloc(1024 * 64)

	streaming := download.NewStreaming(url, size, m.httpBuffer, m)
	m.streaming = streaming

	ioctx := libav.NewAVIOContext(buf, func(buf libav.AVObject) int {
		if buf.Size() == 0 {
			return 0
		}

		if m.httpBuffer.CurrentPos() >= size {
			return libav.AVERROR_EOF
		}

		require := int64(buf.Size())
		got := m.httpBuffer.Read(&buf, require)
		if got < require && !m.httpBuffer.IsFinish() {
			startWaitTime := time.Now()

			for {
				select {
				case <-time.After(20 * time.Millisecond):
				case <-m.quit:
					return libav.AVERROR_INVALIDDATA
				}

				got += m.httpBuffer.Read(&buf, require-got)
				if got >= require || m.httpBuffer.IsFinish() {
					break
				} else {
					if time.Since(startWaitTime) > download.NetworkTimeout {
						pos := m.httpBuffer.CurrentPos()

						log.Print("Streaming timeout Start:", pos)
						url, _, _, _, err := download.GetDownloadInfoN(file, nil, 10, false, m.quit)
						if err != nil {
							m.w.SendAlert(fmt.Sprintf(`Couldn't download "%s"`, file))
							streaming.Stop()
							return libav.AVERROR_INVALIDDATA
						}
						streaming.SetUrl(url)
						startWaitTime = time.Now()
						go streaming.Start(pos, m.quit)
					}
				}
			}
		}

		return int(got)
	}, func(offset int64, whence int) int64 {
		if whence == libav.AVSEEK_SIZE {
			return m.httpBuffer.size
		}

		pos, start := m.httpBuffer.Seek(offset, whence)
		if start >= 0 && start < size {
			go streaming.Start(start, m.quit)
		}
		return pos
	})

	ctx := libav.NewAVFormatContext()
	ctx.SetPb(ioctx)

	m.httpBuffer.Seek(0, os.SEEK_SET)
	go streaming.Start(0, m.quit)

	if err := ctx.OpenInput(file); err != nil {
		return ctx, name, err
	}

	log.Print("open http return")
	return ctx, name, nil
}
