package common

import (
	"container/list"
	"sync"
	"time"
)

type queued struct {
	when  time.Time
	slice []byte
}

type Maker struct {
	get, give chan []byte
}

type RWMaker struct {
	rm, wm *Maker
}

//var MRec *RWMaker = &RWMaker{rm: createMaker(1050), wm: createMaker(1050)}

func (this *RWMaker) RGet(size int) (buf []byte) {
	buf = <-this.rm.get
	/*
	  select{
	  case buf = <- this.rm.get:
	  default:
	  }
	  if buf == nil || cap(buf) < size {
	    buf = make([]byte, size)
	  }
	*/
	return
}

func (this *RWMaker) RGive(buf []byte) {
	this.rm.give <- buf
}

func (this *RWMaker) WGet(size int) (buf []byte) {
	buf = <-this.wm.get
	/*
	  select {
	  case buf = <- this.wm.get:
	  default:
	  }
	  if buf == nil || cap(buf) < size {
	    buf = make([]byte, size)
	  }
	*/
	return
}

func (this *RWMaker) WGive(buf []byte) {
	this.wm.give <- buf
}

type MakeRecycler struct {
	sync.Mutex
	rmm, wmm map[uint32]*Maker
}

//var MRec *MakeRecycler = NewMakeRecycler()

func NewMakeRecycler() *MakeRecycler {
	tmp := &MakeRecycler{
		rmm: make(map[uint32]*Maker),
		wmm: make(map[uint32]*Maker),
	}
	return tmp
}

func createMaker(size uint32) *Maker {
	m := &Maker{get: make(chan []byte), give: make(chan []byte)}
	go func() {
		q := new(list.List)
		//    timeout := time.NewTimer(time.Minute)
		for {
			if q.Len() == 0 {
				q.PushFront(make([]byte, size))
				//        q.PushFront(queued{when: time.Now(), slice: b})
			}
			e := q.Front()
			select {
			case b := <-m.give:
				q.PushFront(b)
				//          q.PushFront(queued{when: time.Now(), slice: b})
			case m.get <- e.Value.([]byte):
				q.Remove(e)
			}

			if q.Len() > 10 {
				back := q.Back()
				q.Remove(back)
				back.Value = nil
			}

			/*
			   case <- timeout.C:
			     for e != nil {
			       next := e.Next()
			       if time.Since(e.Value.(queued).when) > time.Minute {
			         q.Remove(e)
			         e.Value = nil
			       }
			       e = next
			     }
			   }
			   timeout.Reset(time.Minute)
			*/
		}
	}()
	return m
}
