package reusablebytes

import (
	"fmt"
	"testing"
)

func TestLazy(t *testing.T) {
	rb := ReusableBytes{}
	rb.WriteString("NISHIGIUWOSHGIU")
	l := rb.Lazy()
	rb.Anchor()
	rb.WriteString("MILAOGIU")
	l2 := rb.LazyFromAnchor()
	fmt.Println(l.String())
	fmt.Println(l2.String())
}

func TestReusableBytes(t *testing.T) {
	rb := ReusableBytes{}
	rb.WriteString("NISHIGIUWOSHIGIUMILAOGIU")
	rb.Anchor()
	rb.WriteString("WOSHIGIU")
	s := rb.StringFromAnchor()
	fmt.Println(s)
	fmt.Println(rb.String())
}
