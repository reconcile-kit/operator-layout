package app

import (
	"context"
	"github.com/base-cloud-engine/example-agent/pkg/logger"
	"github.com/reconcile-kit/api/resource"
	"github.com/reconcile-kit/controlloop"
	event "github.com/reconcile-kit/redis-informer-provider"
	state "github.com/reconcile-kit/state-manager-provider"
	"net/http"
	"sync"
)

type InitReconciler[T resource.Object[T]] interface {
	controlloop.Reconcile[T]
	SetStorage(storage *controlloop.StorageSet)
}

type Stopped interface {
	Stop()
}

type Application struct {
	shardID                string
	receivers              []controlloop.Receiver
	informerAddress        string
	externalStorageAddress string
	storageSet             *controlloop.StorageSet
	stopped                []Stopped
	logger                 *logger.Logger
}

func New(shardID string, informerAddr, externalStorageAddr string, logger *logger.Logger) *Application {
	storageSet := controlloop.NewStorageSet()
	return &Application{
		shardID:                shardID,
		informerAddress:        informerAddr,
		externalStorageAddress: externalStorageAddr,
		storageSet:             storageSet,
		logger:                 logger,
	}
}

func (a *Application) AddReceiver(receiver controlloop.Receiver) {
	a.receivers = append(a.receivers, receiver)
}
func (a *Application) AddStopped(stopped Stopped) {
	a.stopped = append(a.stopped, stopped)
}

func (a *Application) Stop() {
	stoopedCount := len(a.stopped)
	if stoopedCount == 0 {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(stoopedCount)
	for _, stopped := range a.stopped {
		go func() {
			defer wg.Done()
			stopped.Stop()
		}()
	}

	wg.Wait()
}

func (a *Application) Run(ctx context.Context) error {
	eventProvider, err := event.NewRedisStreamListener(a.informerAddress, a.shardID)
	if err != nil {
		return err
	}

	informer := controlloop.NewStorageInformer(a.shardID, eventProvider, a.receivers)
	err = informer.Run(ctx)
	if err != nil {
		return err
	}
	return nil
}

func SetRemoteClient[T resource.Object[T]](shardID string, app *Application, gk resource.GroupKind) error {
	sm, err := state.NewStateManagerProvider[T](app.externalStorageAddress, &http.Client{})
	if err != nil {
		return err
	}
	rc, err := controlloop.NewRemoteClient[T](shardID, gk, sm)
	if err != nil {
		return err
	}
	controlloop.SetStorage[T](app.storageSet, rc)
	return nil
}

func SetController[T resource.Object[T]](app *Application, gk resource.GroupKind, controller InitReconciler[T], opts ...controlloop.StorageOption) error {
	sm, err := state.NewStateManagerProvider[T](app.externalStorageAddress, &http.Client{})
	if err != nil {
		return err
	}
	sc, err := controlloop.NewStorageController[T](
		app.shardID,
		gk,
		sm,
		controlloop.NewMemoryStorage[T](opts...),
	)
	if err != nil {
		return err
	}

	controlloop.SetStorage[T](app.storageSet, sc)
	controller.SetStorage(app.storageSet)
	currentLoop := controlloop.New[T](controller, sc, controlloop.WithLogger(app.logger))
	currentLoop.Run()
	app.AddReceiver(sc)
	app.AddStopped(currentLoop)

	return nil
}
