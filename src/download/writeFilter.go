package download

import (
	"fmt"
	"io"
	"log"
	"native"
	"os"
	"path/filepath"
	"time"
)

type writeFilter struct {
	basicFilter
	w io.Writer
}

func (wf *writeFilter) active() {
	writeOutput(wf.w, wf.input, wf.output, wf.quit)
}

func writeOutput(w io.Writer, input <-chan *block, output chan *block, quit chan bool) {
	pathErrNotifyTimes := 0
	for {
		select {
		case b, ok := <-input:
			if !ok {
				fmt.Println("close write output")
				close(output)
				return
			}
			for {

				_, err := w.Write(b.data)
				b.data = nil

				if err == nil {
					select {
					case output <- b:
						break
					case <-quit:
						return
					}
					break
				} else if perr, ok := err.(*os.PathError); ok {
					log.Print(err)

					if pathErrNotifyTimes == 0 { //only report once
						native.SendNotification("Error write "+filepath.Base(perr.Path), perr.Err.Error())
					}
					pathErrNotifyTimes++
					if pathErrNotifyTimes > 100 {
						log.Fatal(err)
						return
					}

					select {
					case <-quit:
						return
					case <-time.After(time.Second * 2):
						break
					}
				} else {
					log.Print(err)
					ensureQuit(quit)
					return
				}
			}
		case <-quit:
			fmt.Println("write output quit")
			return

		}
	}

	fmt.Println("writeOutput end")
}
