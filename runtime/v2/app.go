package runtime

import (
	"encoding/json"
	"errors"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"golang.org/x/exp/slices"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/legacy"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
)

// App is a wrapper around AppManager and ModuleManager that can be used in hybrid
// app.go/app config scenarios or directly as a servertypes.Application instance.
// To get an instance of *App, *AppBuilder must be requested as a dependency
// in a container which declares the runtime module and the AppBuilder.Build()
// method must be called.
//
// App can be used to create a hybrid app.go setup where some configuration is
// done declaratively with an app config and the rest of it is done the old way.
// See simapp/app_v2.go for an example of this setup.
type App[T transaction.Tx] struct {
	*appmanager.AppManager[T]

	// app manager dependencies
	stf                *stf.STF[T]
	msgRouterBuilder   *stf.MsgRouterBuilder
	queryRouterBuilder *stf.MsgRouterBuilder
	db                 Store

	// app configuration
	logger    log.Logger
	config    *runtimev2.Module
	appConfig *appv1alpha1.Config

	// modules configuration
	storeKeys          []string
	interfaceRegistrar registry.InterfaceRegistrar
	amino              legacy.Amino
	moduleManager      *MM[T]

	// GRPCQueryDecoders maps gRPC method name to a function that decodes the request
	// bytes into a gogoproto.Message, which then can be passed to appmanager.
	GRPCQueryDecoders map[string]func(requestBytes []byte) (gogoproto.Message, error)
}

// Logger returns the app logger.
func (a *App[T]) Logger() log.Logger {
	return a.logger
}

// ModuleManager returns the module manager.
func (a *App[T]) ModuleManager() *MM[T] {
	return a.moduleManager
}

// DefaultGenesis returns a default genesis from the registered modules.
func (a *App[T]) DefaultGenesis() map[string]json.RawMessage {
	return a.moduleManager.DefaultGenesis()
}

// LoadLatest loads the latest version.
func (a *App[T]) LoadLatest() error {
	return a.db.LoadLatestVersion()
}

// LoadHeight loads a particular height
func (a *App[T]) LoadHeight(height uint64) error {
	return a.db.LoadVersion(height)
}

// Close is called in start cmd to gracefully cleanup resources.
func (a *App[T]) Close() error {
	return nil
}

// GetStoreKeys returns all the app store keys.
func (a *App[T]) GetStoreKeys() []string {
	return a.storeKeys
}

// UnsafeFindStoreKey fetches a registered StoreKey from the App in linear time.
// NOTE: This should only be used in testing.
func (a *App[T]) UnsafeFindStoreKey(storeKey string) (string, error) {
	i := slices.IndexFunc(a.storeKeys, func(s string) bool { return s == storeKey })
	if i == -1 {
		return "", errors.New("store key not found")
	}

	return a.storeKeys[i], nil
}

// GetStore returns the app store.
func (a *App[T]) GetStore() Store {
	return a.db
}

// GetLogger returns the app logger.
func (a *App[T]) GetLogger() log.Logger {
	return a.logger
}

func (a *App[T]) ExecuteGenesisTx(_ []byte) error {
	panic("App.ExecuteGenesisTx not supported in runtime/v2")
}

func (a *App[T]) GetAppManager() *appmanager.AppManager[T] {
	return a.AppManager
}

func (a *App[T]) GetGRPCQueryDecoders() map[string]func(requestBytes []byte) (gogoproto.Message, error) {
	return a.GRPCQueryDecoders
}
