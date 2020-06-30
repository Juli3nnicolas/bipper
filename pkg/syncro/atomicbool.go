package syncro

import "sync/atomic"

// AtomicBool is a struct that enables to manipulate a boolean value in an atomic ("thread-safe")
// manner
type AtomicBool struct {
	boolean uint32
}

// NewAtomicBool creates an atomic bool value initialised
// with the "init" value
func NewAtomicBool(init bool) *AtomicBool {
	b := &AtomicBool{}
	if init == true {
		b.True()
	} else {
		b.False()
	}

	return b
}

// True Sets the atomic boolean to true in an atomic way
func (o *AtomicBool) True() {
	atomic.StoreUint32(&o.boolean, 1)
}

// False sets the atomic boolean to false in an atomic way
func (o *AtomicBool) False() {
	atomic.StoreUint32(&o.boolean, 0)
}

// Value gets the current boolean's value in an atomic way
func (o *AtomicBool) Value() bool {
	return atomic.LoadUint32(&o.boolean) == 1
}
