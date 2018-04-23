package common

import (
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	HEART_BEAT       = 10
	TIMER_QUEUE_SIZE = 1024
	TIME_OUT_FACTOR  = 3
)

var identity uint32 = 0

func InitIdentity(id uint32) {
	identity = id
}

func NewID() uint32 {
	return atomic.AddUint32(&identity, 1)
}

//获取几分之的几率
func SelectByOdds(upNum, downNum uint32) bool {
	if downNum < 1 {
		return false
	}
	if upNum < 1 {
		return false
	}
	if upNum > downNum-1 {
		return true
	}
	return (1 + uint32((float64(rand.Int63())/(1<<63))*float64(downNum))) <= upNum
}

//获取百分之的几率
func SelectByPercent(percent uint32) bool {
	return SelectByOdds(percent, 100)
}

//获取千分之的几率
func SelectByThousand(th uint32) bool {
	return SelectByOdds(th, 1000)
}

//获取万分之的几率
func SelectByTenTh(tenth uint32) bool {
	return SelectByOdds(tenth, 10000)
}

//获取十万分之的几率
func SelectByLakh(lakh uint32) bool {
	return SelectByOdds(lakh, 100000)
}

func DailyZero() uint32 {
	now := time.Now()
	hour, minute, second := now.Hour(), now.Minute(), now.Second()
	return uint32(now.Unix() - int64((hour*3600)+(minute*60)+second))
}

func DailyHour() uint32 {
	now := time.Now()
	minute, second := now.Minute(), now.Second()
	return uint32(now.Unix() - int64((minute*60)+second))
}

func DailyZeroByTime(ts uint32) uint32 {
	t := time.Unix(int64(ts), 0)
	hour, minute, second := t.Hour(), t.Minute(), t.Second()
	return uint32(int64(ts) - int64((hour*3600)+(minute*60)+second))
}

func UniqueWeek(tm time.Time) uint32 {
	year, week := tm.ISOWeek()
	return (uint32(year) << 16) + uint32(week)
}

// Ascii numbers 0-9
const (
	ascii_0 = 48
	ascii_9 = 57
)

func ParseUint64(d []byte) (uint64, bool) {
	var n uint64
	d_len := len(d)
	if d_len == 0 {
		return 0, false
	}
	for i := 0; i < d_len; i++ {
		j := d[i]
		if j < ascii_0 || j > ascii_9 {
			return 0, false
		}
		n = n*10 + (uint64(j - ascii_0))
	}
	return n, true
}

func IpV4ToUint32(ip string) uint32 {
	var n uint32
	ips := strings.Split(ip, ".")
	if len(ips) != 4 {
		return n
	}
	b0, _ := strconv.Atoi(ips[0])
	b1, _ := strconv.Atoi(ips[1])
	b2, _ := strconv.Atoi(ips[2])
	b3, _ := strconv.Atoi(ips[3])
	n += uint32(b0) << 24
	n += uint32(b1) << 16
	n += uint32(b2) << 8
	n += uint32(b3)
	return n
}

func StringToUint32Slice(s string, seq string) (ret []uint32) {
	if len(s) == 0 {
		return
	}
	set := strings.Split(s, seq)
	ret = make([]uint32, len(set))
	for index, value := range set {
		tmp, _ := strconv.ParseUint(value, 10, 32)
		ret[index] = uint32(tmp)
	}
	return
}

func Uint32SliceToString(set []uint32, seq string) (ret string) {
	set_len := len(set)
	if set_len == 0 {
		return
	}
	for index, value := range set {
		ret += strconv.FormatUint(uint64(value), 10)
		if index < set_len-1 {
			ret += seq
		}
	}
	return
}
