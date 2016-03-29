package fsm

// For something to be managed, it has to have a key or
// indexable attribute.
type Managed interface {
	GetKey() interface{}
}

// Managed set of objects by some common criteria / key
type ManagedSet interface {
	Managed
	Add(Managed) *Fsm
	Remove(Managed)
	Instances() []*Fsm
	Empty() bool
}
