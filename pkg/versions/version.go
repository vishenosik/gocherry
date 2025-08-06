package versions

import "fmt"

type Interface interface {
	fmt.Stringer
	In_(v1, v2 Interface) bool
}
