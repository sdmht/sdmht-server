package graph

import (
	"io"
)

type Void struct{}

func (Void) MarshalGQL(w io.Writer) {
	w.Write([]byte("null"))
}
func (*Void) UnmarshalGQL(v any) error {
	return nil
}
