package api

import (
	"github.com/reconcile-kit/api/conditions"
	"github.com/reconcile-kit/api/resource"
)

type ExampleResource struct {
	resource.Resource
	Spec   ExampleResourceSpec   `json:"spec"`
	Status ExampleResourceStatus `json:"status"`
}
type ExampleResourceSpec struct {
}

type ExampleResourceStatus struct {
	Conditions []conditions.Condition `json:"conditions"`
}

func (c *ExampleResource) GetConditions() []conditions.Condition {
	return c.Status.Conditions
}

func (c *ExampleResource) SetConditions(i []conditions.Condition) {
	c.Status.Conditions = i
}

func (c *ExampleResource) DeepCopy() *ExampleResource {
	return resource.DeepCopyStruct(c).(*ExampleResource)
}
