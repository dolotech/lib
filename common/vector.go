package common

import (
	"sync"
)

type VectorFun func(interface{}, interface{}) bool

type Op func(interface{})

type Vector struct {
	Values []interface{}
	len    int
	equal  VectorFun
	sync.RWMutex
}

func NewVector(vfunc VectorFun) *Vector {
	return &Vector{len: 0, equal: vfunc}
}

func (this *Vector) Len() int {
	this.RLock()
	defer this.RUnlock()
	return this.len
}

func (this *Vector) Index(i int) interface{} {
	this.RLock()
	defer this.RUnlock()
	if i >= this.len {
		return nil
	}
	return this.Values[i]
}

func (this *Vector) expand() {
	curcap := len(this.Values)
	var newcap int
	if curcap == 0 {
		newcap = 8
	} else if curcap < 1024 {
		newcap = curcap * 2
	} else {
		newcap = curcap + (curcap / 4)
	}
	values := make([]interface{}, newcap)
	if curcap != 0 {
		copy(values, this.Values)
	}
	this.Values = values
}

func (this *Vector) PushBack(value interface{}) {
	this.Lock()
	defer this.Unlock()
	if len(this.Values) == this.len {
		this.expand()
	}
	this.Values[this.len] = value
	this.len++
}

func (this *Vector) PopBack() interface{} {
	this.Lock()
	defer this.Unlock()
	if this.len == 0 {
		return nil
	}
	this.len--
	return this.Values[this.len]
}

func (this *Vector) Remove(value interface{}) {
	this.Lock()
	defer this.Unlock()
	for i := 0; i < this.len; {
		if this.equal(this.Values[i], value) == true {
			if tmp := i + 1; tmp < this.len {
				copy(this.Values[i:], this.Values[tmp:this.len])
			}
			this.len--
		} else {
			i++
		}
	}
}

func (this *Vector) Traverse(op Op) {
	this.Lock()
	defer this.Unlock()
	for i := 0; i < this.len; i++ {
		op(this.Values[i])
	}
}
