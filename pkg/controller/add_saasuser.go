package controller

import (
	"github.com/redhat-developer/saas-next/pkg/controller/saasuser"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, saasuser.Add)
}
