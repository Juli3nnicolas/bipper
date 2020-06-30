package syncro

import "sync/atomic"

type AtomicBool struct {
	boolean uint32
}

func NewAtomicBool(init bool) *AtomicBool {
	b := &AtomicBool{}
	if init == true {
		b.True()
	} else {
		b.False()
	}

	return b
}

func (o *AtomicBool) True() {
	atomic.StoreUint32(&o.boolean, 1)
}

func (o *AtomicBool) False() {
	atomic.StoreUint32(&o.boolean, 0)
}

func (o *AtomicBool) Value() bool {
	return atomic.LoadUint32(&o.boolean) == 1
}
