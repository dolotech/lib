package common

/* 根据chan bool设计Mutex的依据
1:  A send on a channel happens before the corresponding receive from that channel completes.
2:  The closing of a channel happens before a receive that returns a zero value because the channel is closed.
3:  A receive from an unbuffered channel happens before the send on that channel completes.
*/

type lock chan bool

func newLock() lock    { l := make(lock, 1); l <- true; return l }
func (l lock) Lock()   { <-l }
func (l lock) Unlock() { l <- true }
