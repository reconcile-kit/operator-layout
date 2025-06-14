package controllers

import (
	"context"
	"github.com/base-cloud-engine/example-agent/api"
	"github.com/base-cloud-engine/example-agent/internal/services/exampleresource"
	"github.com/base-cloud-engine/example-agent/pkg/logger"
	"github.com/reconcile-kit/api/conditions"
	"github.com/reconcile-kit/api/resource"
	cl "github.com/reconcile-kit/controlloop"
	"slices"
)

const exampleResourceFinalizer = "example-resource.central.base-cloud-engine.com/finalizer"

type ExampleResourceReconciler[T resource.Object[T]] struct {
	exampleResourceService *exampleresource.Service
	storage                *cl.StorageSet
	logger                 *logger.Logger
}

func (r *ExampleResourceReconciler[T]) Reconcile(ctx context.Context, object *api.ExampleResource) (result cl.Result, reterr error) {
	exampleResourceClient, ok := cl.GetStorage[*api.ExampleResource](r.storage)
	if !ok {
		return cl.Result{}, nil
	}
	defer func() {
		emptyResult := cl.Result{}
		if reterr != nil {
			r.logger.Error(reterr)
		}
		if result == emptyResult && reterr == nil {
			object.SetCurrentVersion(object.GetVersion())
		}
		conditions.SyncReady(object)
		err := exampleResourceClient.UpdateStatus(ctx, object)
		if err != nil {
			reterr = err
			r.logger.Errorf("Error update helper %s", err)
		}
	}()

	if !slices.Contains(object.Finalizers, exampleResourceFinalizer) && object.DeletionTimestamp == "" {
		object.Finalizers = append(object.Finalizers, exampleResourceFinalizer)
		return cl.Result{Requeue: true}, nil
	}

	if object.DeletionTimestamp != "" {
		return r.reconcileDelete(ctx, object)
	}

	return r.reconcileNormal(ctx, object)
}

func (r *ExampleResourceReconciler[T]) reconcileNormal(ctx context.Context, object *api.ExampleResource) (cl.Result, error) {
	conditions.MarkTrue(object, api.RunMainServerCond)

	return cl.Result{}, nil
}

func (r *ExampleResourceReconciler[T]) reconcileDelete(ctx context.Context, object *api.ExampleResource) (cl.Result, error) {
	r.logger.Infof("Start reconcile delete %s %s", object.Namespace, object.Name)

	object.Finalizers = []string{}
	r.logger.Infof("finalizer removed %s %s", object.Namespace, object.Name)
	return cl.Result{}, nil
}

func NewExampleResourceReconciler[T resource.Object[T]](l *logger.Logger, exampleResourceService *exampleresource.Service) *ExampleResourceReconciler[T] {
	return &ExampleResourceReconciler[T]{
		logger:                 l,
		exampleResourceService: exampleResourceService,
	}
}
func (r *ExampleResourceReconciler[T]) SetStorage(storage *cl.StorageSet) {
	r.storage = storage
}
