package skiplist

import (
	"fmt"
	"math/rand"
)

var SKIPLIST_P float32 = 0.25

const SKIPLIST_MAXLEVEL int = 32

type SkipListLevel struct {
	forward *SkipListNode
	span    uint32
}

type SkipListNode struct {
	value    interface{}
	backward *SkipListNode
	level    []SkipListLevel
}

func NewSkipListNode(level int, value interface{}) *SkipListNode {
	sln := &SkipListNode{
		value: value,
		level: make([]SkipListLevel, level),
	}
	return sln
}

func (this *SkipListNode) Next(i int) *SkipListNode {
	return this.level[i].forward
}

func (this *SkipListNode) SetNext(i int, next *SkipListNode) {
	this.level[i].forward = next
}

func (this *SkipListNode) Span(i int) uint32 {
	return this.level[i].span
}

func (this *SkipListNode) SetSpan(i int, span uint32) {
	this.level[i].span = span
}

func (this *SkipListNode) Value() interface{} {
	return this.value
}

func (this *SkipListNode) Prev() *SkipListNode {
	return this.backward
}

type Comparatorer interface {
	CmpScore(interface{}, interface{}) int
	CmpKey(interface{}, interface{}) int
}

type SkipList struct {
	head, tail    *SkipListNode
	length, level uint32
	Comparatorer
}

func NewSkipList(cmp Comparatorer) *SkipList {
	sl := &SkipList{
		level:        1,
		length:       0,
		tail:         nil,
		Comparatorer: cmp,
	}
	sl.head = NewSkipListNode(SKIPLIST_MAXLEVEL, nil)
	for i := 0; i < SKIPLIST_MAXLEVEL; i++ {
		sl.head.SetNext(i, nil)
		sl.head.SetSpan(i, 0)
	}
	sl.head.backward = nil
	return sl
}

func (this *SkipList) Level() uint32 { return this.level }

func (this *SkipList) Length() uint32 { return this.length }

func (this *SkipList) Head() *SkipListNode { return this.head }

func (this *SkipList) Tail() *SkipListNode { return this.tail }

func (this *SkipList) First() *SkipListNode { return this.head.Next(0) }

func (this *SkipList) randomLevel() int {
	level := 1
	for (rand.Uint32()&0xFFFF) < uint32(SKIPLIST_P*0xFFFF) && level < SKIPLIST_MAXLEVEL {
		level++
	}
	return level
}

func (this *SkipList) Insert(value interface{}) *SkipListNode {
	var update [SKIPLIST_MAXLEVEL]*SkipListNode
	var rank [SKIPLIST_MAXLEVEL]uint32
	x := this.head
	for i := int(this.level - 1); i >= 0; i-- {
		if i == int(this.level-1) {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for next := x.Next(i); next != nil &&
			(this.CmpScore(next.value, value) < 0 ||
				(this.CmpScore(next.value, value) == 0 &&
					this.CmpKey(next.value, value) < 0)); next = x.Next(i) {
			rank[i] += x.Span(i)
			x = next
		}
		update[i] = x
	}

	level := uint32(this.randomLevel())

	if level > this.level {
		for i := this.level; i < level; i++ {
			rank[i] = 0
			update[i] = this.head
			update[i].SetSpan(int(i), this.length)
		}
		this.level = level
	}

	x = NewSkipListNode(int(level), value)
	for i := 0; i < int(level); i++ {
		x.SetNext(i, update[i].Next(i))
		update[i].SetNext(i, x)

		x.SetSpan(i, update[i].Span(i)-(rank[0]-rank[i]))
		update[i].SetSpan(i, rank[0]-rank[i]+1)
	}

	for i := level; i < this.level; i++ {
		update[i].SetSpan(int(i), update[i].Span(int(i))+1)
	}

	if update[0] == this.head {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	if x.Next(0) != nil {
		x.Next(0).backward = x
	} else {
		this.tail = x
	}
	this.length++
	return x
}

func (this *SkipList) DeleteNode(x *SkipListNode, update []*SkipListNode) {
	for i := 0; i < int(this.level); i++ {
		if update[i].Next(i) == x {
			update[i].SetSpan(i, update[i].Span(i)+x.Span(i)-1)
			update[i].SetNext(i, x.Next(i))
		} else {
			update[i].SetSpan(i, update[i].Span(i)-1)
		}
	}

	if x.Next(0) != nil {
		x.Next(0).backward = x.backward
	} else {
		this.tail = x.backward
	}

	for this.level > 1 && this.head.Next(int(this.level-1)) == nil {
		this.level--
	}
	this.length--
}

func (this *SkipList) Delete(value interface{}) int {
	update := make([]*SkipListNode, int(this.level))
	var x *SkipListNode = this.head
	for i := int(this.level - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			(this.CmpScore(next.value, value) < 0 ||
				(this.CmpScore(next.value, value) == 0 &&
					this.CmpKey(next.value, value) < 0)); next = x.Next(i) {
			x = next
		}
		update[i] = x
	}

	x = x.Next(0)
	if x != nil &&
		this.CmpKey(x.value, value) == 0 &&
		this.CmpScore(x.value, value) == 0 {
		this.DeleteNode(x, update)
		return 1
	}
	return 0
}

//TODO: 1-based rank
func (this *SkipList) GetRank(value interface{}) uint32 {
	var rank uint32 = 0
	x := this.head
	for i := int(this.level - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			(this.CmpScore(next.value, value) < 0 ||
				(this.CmpScore(next.value, value) == 0 &&
					this.CmpKey(next.value, value) <= 0)); next = x.Next(i) {
			rank += x.Span(i)
			x = next
		}
		if x != this.head && this.CmpKey(x.value, value) == 0 {
			return rank
		}
	}
	return 0
}

func (this *SkipList) GetNodeByRank(rank uint32) *SkipListNode {
	x := this.head
	var traversed uint32 = 0
	for i := int(this.level - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			traversed+x.Span(i) <= rank; next = x.Next(i) {
			traversed += x.Span(i)
			x = next
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

func (this *SkipList) Dump() {
	fmt.Println("*************SKIP LIST DUMP START*************")
	for i := int(this.level - 1); i >= 0; i-- {
		fmt.Printf("level:--------%v--------\n", i)
		x := this.head
		for x != nil {
			if x == this.head {
				fmt.Printf("Head span: %v\n", x.Span(i))
			} else {
				fmt.Printf("span: %v value : %v\n", x.Span(i), x.Value())
			}
			x = x.Next(i)
		}
	}
	fmt.Println("*************SKIP LIST DUMP END*************")
}
