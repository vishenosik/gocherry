package versions

import "fmt"

type Interface interface {
	fmt.Stringer
	Parse_(string) (Interface, error)
	In_(v1, v2 Interface) bool
}
