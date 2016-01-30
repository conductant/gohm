package registry

type Action int

type Trigger interface {
	Event() <-chan interface{}
}

type Delete struct {
	Path `json:"path"`
}

type Create struct {
	Path `json:"path"`
}

type Change struct {
	Path `json:"path"`
}

// For equality, set both min and max.  For not equals, set min, max and OutsideRange to true.
type Members struct {
	Path         `json:"path"`
	Min          *int `json:"min,omitempty"`
	Max          *int `json:"max,omitempty"`
	Delta        *int `json:"delta,omitempty"`         // delta of count
	OutsideRange bool `json:"outside_range,omitempty"` // default is within range.  true for outside range.
}

func (this *Members) SetMin(min int) *Members {
	this.Min = &min
	return this
}

func (this *Members) SetMax(max int) *Members {
	this.Max = &max
	return this
}

func (this *Members) SetDelta(d int) *Members {
	this.Delta = &d
	return this
}

func (this *Members) SetOutsideRange(b bool) *Members {
	this.OutsideRange = b
	return this
}
