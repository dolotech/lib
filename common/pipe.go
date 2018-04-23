package common

import "container/list"

func Pipe(in chan []byte, out chan []byte) {
	var msg []byte
	var outCh chan []byte
	q := list.New()

	for {
		select {
		case info := <-in:
			q.PushBack(info)
			if msg == nil {
				msg = q.Remove(q.Front()).([]byte)
				outCh = out
			}
		case outCh <- msg:
			if q.Len() > 0 {
				msg = q.Remove(q.Front()).([]byte)
			} else {
				msg = nil
				outCh = nil
			}
		}
	}
}
