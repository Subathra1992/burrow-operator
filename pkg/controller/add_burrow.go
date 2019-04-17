package controller

import (
	"github.com/subravi92/burrow-operator/pkg/controller/burrow"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, burrow.Add)
}
