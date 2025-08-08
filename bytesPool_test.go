package reusablebytes

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	bp := new(BytesPool)
	bp.Init(1, 131072)
	for i := 0; i < 128; i++ {
		rb, id := bp.Get()
		fmt.Println(rb, " ", id)
		rb.WriteString(fmt.Sprintf("NISHIGIU%d", i))
		fmt.Println(rb.String())
	}
	for i := 0; i < 128; i++ {
		bp.Put(int32(i))
	}
}
