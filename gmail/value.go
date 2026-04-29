package gmail

import "encoding/json"

type Value[T any] struct {
	Src string
	Err error
	Val T
}

func (v Value[T]) MashalJSON() ([]byte, error) {
	jn := jsonValue[T]{
		Src: v.Src,
		Err: v.Err.Error(),
		Val: v.Val,
	}
	return json.MarshalIndent(jn, "", "  ")

}

type jsonValue[T any] struct {
	Src string `json:"src"`
	Err string `json:"err,omitempty"`
	Val T      `json:"val,omitempty"`
}
