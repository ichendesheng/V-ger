package movie

import (
	"download"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	// "path/filepath"
	. "player/libav"
	"task"
	"time"
	// "task"
	// "time"
	"util"
)

func downloadBytes(url string, from int64, size int, filesize int64) []byte {
	to := from + int64(size)
	if to > filesize {
		to = 0
	}

	println("request:", from, to)
	req := download.CreateDownloadRequest(url, from, to-1)
	resp, _ := http.DefaultClient.Do(req)

	data, _ := ioutil.ReadAll(resp.Body)
	println("get:", len(data), from, size)
	return data
}
func max(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func (m *Movie) openHttp(file string) (AVFormatContext, string) {
	download.NetworkTimeout = 15 * time.Second
	download.BaseDir = util.ReadConfig("dir")

	_, name, size, err := download.GetDownloadInfo(file)

	if err != nil {
		log.Fatal(err)
	}

	t, err := task.GetTask(name)
	if err != nil {
		t = &task.Task{}
		t.Name = name
		t.Size = size
		t.StartTime = time.Now().Unix()
		t.Status = "Playing"
		t.URL = file
		task.SaveTask(t)
	} else {
		t.Status = "Playing"
		task.SaveTask(t)
	}

	m.httpBuffer = NewBuffer(size)

	buf := AVObject{}
	buf.Malloc(1024 * 64)
	ioctx := NewAVIOContext(buf, func(buf AVObject) int {
		if buf.Size() == 0 {
			return 0
		}
		require := int64(buf.Size())

		got := m.httpBuffer.Read(&buf, require)

		if got < require && !m.httpBuffer.IsFinish() {
			if m.c != nil {
				m.c.Pause()
				defer m.c.Resume()
				// defer m.w.SendHideMessage()

				// go m.w.SendShowMessage("Bufferring...", false)
				// m.httpBuffer.Wait(max(require-got, 2*1024*1024))
			}

			for got < require && !m.httpBuffer.IsFinish() {
				time.Sleep(100 * time.Millisecond)
				got += m.httpBuffer.Read(&buf, require-got)
			}
		}

		return int(got)
	}, func(offset int64, whence int) int64 {
		println("seek:", offset, whence)
		if whence == AVSEEK_SIZE {
			return m.httpBuffer.size
		}

		pos, start := m.httpBuffer.Seek(offset, whence)
		if start >= 0 && start < size {
			go func() {
				t, err := task.GetTask(name)
				if err != nil {
					log.Fatal(err)
				}

				download.Streaming(t, m.httpBuffer, start, m.p)
			}()
		}
		return pos
	})

	ctx := NewAVFormatContext()
	ctx.SetPb(ioctx)

	go download.Streaming(t, m.httpBuffer, 0, nil)
	m.httpBuffer.Seek(0, os.SEEK_SET)

	ctx.OpenInput(name)

	println("open http return")
	return ctx, name
}
