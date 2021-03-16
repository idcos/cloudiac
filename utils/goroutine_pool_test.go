package utils

import (
	"fmt"
	"testing"
)

func TestGoroutinePool(t *testing.T) {
	gp := NewGoroutinePool(1)
	names := []string{"1", "2"}
	for i := range names {
		fmt.Printf("go run '%v'\n", i)
		idx := i
		gp.Run(func() {
			fmt.Printf("run '%v'\n", idx)
		})
	}
	gp.Wait()
}
