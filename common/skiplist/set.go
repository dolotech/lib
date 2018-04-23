package skiplist

type Valuer interface {
	Key() uint64
	Score() uint64
}

type Set struct {
	sl    *SkipList
	index map[uint64]Valuer
}

func NewSet(cmp Comparatorer) *Set {
	return &Set{
		sl:    NewSkipList(cmp),
		index: make(map[uint64]Valuer),
	}
}

func (this *Set) Length() uint32 { return this.sl.Length() }

func (this *Set) Head() *SkipListNode { return this.sl.Head() }

func (this *Set) Tail() *SkipListNode { return this.sl.Tail() }

func (this *Set) First() *SkipListNode { return this.sl.First() }

//Insert之前检查element是否存在(GetElement)
func (this *Set) Insert(value Valuer) {
	this.sl.Insert(value)
	this.index[value.Key()] = value
}

func (this *Set) GetElement(key uint64) Valuer {
	if value, exist := this.index[key]; exist {
		return value
	}
	return nil
}

func (this *Set) Delete(value Valuer) {
	delete(this.index, value.Key())
	this.sl.Delete(value)
}

func (this *Set) GetRank(key uint64) uint32 {
	if value, exist := this.index[key]; exist {
		return this.sl.GetRank(value)
	}
	return 0
}

func (this *Set) GetNodeByRank(rank uint32) *SkipListNode {
	return this.sl.GetNodeByRank(rank)
}

func (this *Set) DeleteRangeByRank(start, end uint32) uint32 {
	level := int(this.sl.Level())
	update := make([]*SkipListNode, level)
	var removed uint32 = 0
	var traversed uint32 = 0
	x := this.sl.Head()
	for i := level - 1; i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			x.Span(i)+traversed < start; next = x.Next(i) {
			traversed += x.Span(i)
			x = next
		}
		update[i] = x
	}
	x = x.Next(0)
	traversed++
	for x != nil && traversed <= end {
		next := x.Next(0)
		this.sl.DeleteNode(x, update)
		delete(this.index, x.Value().(Valuer).Key())
		removed++
		traversed++
		x = next
	}
	return removed
}

func (this *Set) Dump() {
	this.sl.Dump()
}

//1-based rank
func (this *Set) GetRightRange(start, end uint32, reversal bool) (uint32, uint32) {
	length := this.sl.Length()
	if length == 0 || start == 0 || end < start || start > length {
		return 0, 0
	}
	if reversal {
		start = length + 1 - start
		if end > length {
			end = 1
		} else {
			end = length + 1 - end
		}
	} else {
		if end > length {
			end = length
		}
	}
	return start, end
}

type RangeSpec struct {
	MinEx, MaxEx bool
	Min, Max     uint64
}

func (this *Set) ValueGteMin(value uint64, spec *RangeSpec) bool {
	if spec.MinEx {
		return value > spec.Min
	}
	return value >= spec.Min
}

func (this *Set) ValueLteMax(value uint64, spec *RangeSpec) bool {
	if spec.MaxEx {
		return value < spec.Max
	}
	return value <= spec.Max
}

func (this *Set) IsInRange(rg *RangeSpec) bool {
	if rg.Min > rg.Max ||
		(rg.Min == rg.Max && (rg.MinEx || rg.MaxEx)) {
		return false
	}

	x := this.sl.Tail()
	if x == nil || !this.ValueGteMin(x.Value().(Valuer).Score(), rg) {
		return false
	}

	x = this.sl.First()
	if x == nil || !this.ValueLteMax(x.Value().(Valuer).Score(), rg) {
		return false
	}
	return true
}

func (this *Set) FirstInRange(rg *RangeSpec) *SkipListNode {
	if !this.IsInRange(rg) {
		return nil
	}

	x := this.sl.Head()
	for i := int(this.sl.Level() - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			!this.ValueGteMin(next.Value().(Valuer).Score(), rg); next = x.Next(i) {
			x = next
		}
	}
	x = x.Next(0)
	if !this.ValueLteMax(x.Value().(Valuer).Score(), rg) {
		return nil
	}
	return x
}

func (this *Set) LastInRange(rg *RangeSpec) *SkipListNode {
	if !this.IsInRange(rg) {
		return nil
	}

	x := this.sl.Head()
	for i := int(this.sl.Level() - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			this.ValueLteMax(next.Value().(Valuer).Score(), rg); next = x.Next(i) {
			x = next
		}
	}
	if !this.ValueGteMin(x.Value().(Valuer).Score(), rg) {
		return nil
	}
	return x
}

func (this *Set) DeleteRangeByScore(rg *RangeSpec) uint32 {
	update := make([]*SkipListNode, int(this.sl.Level()))
	var removed uint32 = 0
	x := this.sl.Head()
	for i := int(this.sl.Level() - 1); i >= 0; i-- {
		for next := x.Next(i); next != nil &&
			((rg.MinEx && next.Value().(Valuer).Score() <= rg.Min) ||
				(!rg.MinEx && next.Value().(Valuer).Score() < rg.Min)); next = x.Next(i) {
			x = next
		}
		update[i] = x
	}
	x = x.Next(0)
	for x != nil &&
		((rg.MaxEx && x.Value().(Valuer).Score() < rg.Max) ||
			(!rg.MaxEx && x.Value().(Valuer).Score() <= rg.Max)) {
		next := x.Next(0)
		this.sl.DeleteNode(x, update)
		delete(this.index, x.Value().(Valuer).Key())
		removed++
		x = next
	}
	return removed
}

func (this *Set) GetRangeByScore(rg *RangeSpec) []interface{} {
	var values []interface{}
	x := this.FirstInRange(rg)
	for x != nil {
		if !this.ValueLteMax(x.Value().(Valuer).Score(), rg) {
			break
		}
		values = append(values, x.value)
		x = x.Next(0)
	}
	return values
}

func (this *Set) Range(f func(interface{})) {
	for tmp := this.First(); tmp != nil; tmp = tmp.Next(0) {
		f(tmp.Value())
	}
}
