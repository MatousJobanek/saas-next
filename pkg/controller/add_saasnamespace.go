package controller

import (
	"github.com/redhat-developer/saas-next/pkg/controller/saasnamespace"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, saasnamespace.Add)
}
