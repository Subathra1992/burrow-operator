package burrow

import "fmt"

type burrowerror struct {
	error string
}

func (e *burrowerror) InvalidNamespaceError() string {
	return fmt.Sprintf("Invalid Namespace.You are not allowed to create in %v ", e.error)
}
