package common

import (
	"sync"
	"time"
)

type HeartBeater interface {
	HeartBeat()
	TimeOut()
}

type Sessioner interface {
	HeartBeater
	RemoteAddr() string
}

type HeartBeat struct {
	keep int
	hb   *time.Timer
	to   *time.Timer
}

func NewHeartBeat(kp int, session HeartBeater) *HeartBeat {
	if kp == 0 {
		kp = HEART_BEAT
	}
	return &HeartBeat{
		keep: kp,
		hb:   time.AfterFunc(time.Second*time.Duration(kp), func() { session.HeartBeat() }),
		to:   time.AfterFunc(time.Second*time.Duration(kp)*TIME_OUT_FACTOR, func() { session.TimeOut() }),
	}
}

func (this *HeartBeat) Reset() {
	this.hb.Reset(time.Second * time.Duration(this.keep))
	this.to.Reset(time.Second * time.Duration(this.keep) * TIME_OUT_FACTOR)
}

func (this *HeartBeat) Stop() {
	this.hb.Stop()
	this.to.Stop()
}

type HeartBeatM struct {
	sync.Mutex
	hbm map[string]*HeartBeat
}

func NewHeartBeatM() *HeartBeatM {
	return &HeartBeatM{hbm: make(map[string]*HeartBeat)}
}

func (this *HeartBeatM) Register(kp int, session Sessioner) {
	this.Lock()
	if session != nil {
		if v, exist := this.hbm[session.RemoteAddr()]; exist {
			v.Reset()
		} else {
			this.hbm[session.RemoteAddr()] = NewHeartBeat(kp, session)
		}
	}
	this.Unlock()
}

func (this *HeartBeatM) Stop(session Sessioner) {
	this.Lock()
	if session != nil {
		if v, exist := this.hbm[session.RemoteAddr()]; exist {
			v.Stop()
		}
	}
	this.Unlock()
}
