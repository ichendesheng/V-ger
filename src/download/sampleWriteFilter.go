package download

import "fmt"

type sampleWriteFilter struct {
	basicFilter
	w WriterAtQuit
}

func (swf *sampleWriteFilter) active() {
	for {
		select {
		case b, ok := <-swf.input:
			if !ok {
				fmt.Println("close sample write output")
				close(swf.output)
				return
			}

			swf.w.WriteAtQuit(b.data, b.from, swf.quit)

			swf.writeOutput(b)
			break
		case <-swf.quit:
			fmt.Println("sample write output quit")
			return
		}
	}

	fmt.Println("sampleWriteOutput end")
}
