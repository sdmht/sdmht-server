package graph

import (
	"io"
)

type Void struct{}

func (_ Void) MarshalGQL(w io.Writer) {
	w.Write([]byte("null"))
}
func (_ *Void) UnmarshalGQL(v any) error {
	return nil
}
