package metric

import (
	"strings"
)

// FamilyInterface interface for a family
type FamilyInterface interface {
	Inspect(inspect func(Family))
	ByteSlice() []byte
}

// Family represents a set of metrics with the same name and help text.
type Family struct {
	Name    string
	Type    Type
	Metrics []*Metric
}

// Inspect use to inspect the inside of a Family
func (f Family) Inspect(inspect func(Family)) {
	inspect(f)
}

// ByteSlice returns the given Family in its string representation.
func (f Family) ByteSlice() []byte {
	b := strings.Builder{}
	for _, m := range f.Metrics {
		b.WriteString(f.Name)
		m.Write(&b)
	}

	return []byte(b.String())
}
