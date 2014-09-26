// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package cloudsigma

import (
	"fmt"
	"sync"

	"github.com/Altoros/gosigma"
	"github.com/juju/juju/constraints"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/config"
	"github.com/juju/juju/environs/simplestreams"
	"github.com/juju/juju/environs/storage"
	"github.com/juju/juju/environs/cloudinit"
	"github.com/juju/juju/instance"
	"github.com/juju/juju/provider/common"
)

const (
	CloudsigmaCloudImagesURLTemplate = "https://%v.cloudsigma.com/"
)

// This file contains the core of the Environ implementation.
type environ struct {
	name string

	lock sync.Mutex

	ecfg    *environConfig
	client  *environClient
	storage *environStorage
}

var _ environs.Environ = (*environ)(nil)
var _ simplestreams.HasRegion = (*environ)(nil)

// Name returns the Environ's name.
func (env environ) Name() string {
	return env.name
}

// Provider returns the EnvironProvider that created this Environ.
func (environ) Provider() environs.EnvironProvider {
	return providerInstance
}

// SetConfig updates the Environ's configuration.
//
// Calls to SetConfig do not affect the configuration of values previously obtained
// from Storage.
func (env *environ) SetConfig(cfg *config.Config) error {
	env.lock.Lock()
	defer env.lock.Unlock()

	ecfg, err := validateConfig(cfg, env.ecfg)
	if err != nil {
		return err
	}

	if env.client == nil || env.ecfg == nil || env.ecfg.clientConfigChanged(ecfg) {
		client, err := newClient(ecfg)
		if err != nil {
			return err
		}

		storage, err := newStorage(ecfg, client)
		if err != nil {
			return err
		}

		env.client = client
		env.storage = storage
	}

	env.ecfg = ecfg

	return nil
}

// Config returns the configuration data with which the Environ was created.
// Note that this is not necessarily current; the canonical location
// for the configuration data is stored in the state.
func (env *environ) Config() *config.Config {
	return env.ecfg.Config
}

// Storage returns storage specific to the environment.
func (env environ) Storage() storage.Storage {
	return env.storage
}

// Bootstrap initializes the state for the environment, possibly
// starting one or more instances.  If the configuration's
// AdminSecret is non-empty, the administrator password on the
// newly bootstrapped state will be set to a hash of it (see
// utils.PasswordHash), When first connecting to the
// environment via the juju package, the password hash will be
// automatically replaced by the real password.
//
// The supplied constraints are used to choose the initial instance
// specification, and will be stored in the new environment's state.
//
// Bootstrap is responsible for selecting the appropriate tools,
// and setting the agent-version configuration attribute prior to
// bootstrapping the environment.
func (env *environ) Bootstrap(ctx environs.BootstrapContext, params environs.BootstrapParams) (arch, series string, finalizer environs.BootstrapFinalizer, err error) {
	arch, series, finalizer, err = common.Bootstrap(ctx, env, params)

	if err != nil {
		return
	}

	newFinalizer := func(ctx environs.BootstrapContext, mcfg *cloudinit.MachineConfig) (err error) {
		err = finalizer(ctx, mcfg)
		if err != nil {
			return err
		}

		// provide additional agent config for localstorage, if any
		if env.storage.tmp {
			_, addr, ok := env.client.stateServerAddress()
			if !ok {
				return fmt.Errorf("Can't obtain state server address")
			}
			if err := env.prepareStorage(addr, mcfg); err != nil {
				return fmt.Errorf("failed prepare storage: %v", err)
			}
		}

		return err
	}

	return arch, series, newFinalizer, err
}

func (e *environ) StateServerInstances() ([]instance.Id, error) {
	return common.ProviderStateInstances(e, e.Storage())
}

// Destroy shuts down all known machines and destroys the
// rest of the environment. Note that on some providers,
// very recently started instances may not be destroyed
// because they are not yet visible.
//
// When Destroy has been called, any Environ referring to the
// same remote environment may become invalid
func (env *environ) Destroy() error {
	// You can probably ignore this method; the common implementation should work.
	return common.Destroy(env)
}

// PrecheckInstance performs a preflight check on the specified
// series and constraints, ensuring that they are possibly valid for
// creating an instance in this environment.
//
// PrecheckInstance is best effort, and not guaranteed to eliminate
// all invalid parameters. If PrecheckInstance returns nil, it is not
// guaranteed that the constraints are valid; if a non-nil error is
// returned, then the constraints are definitely invalid.
func (env *environ) PrecheckInstance(series string, cons constraints.Value, placement string) error {
	logger.Infof("cloudsigma:environ:PrecheckInstance")
	return nil
}

// Region is specified in the HasRegion interface.
func (env *environ) Region() (simplestreams.CloudSpec, error) {
	return env.cloudSpec(env.ecfg.region())
}

func (env *environ) cloudSpec(region string) (simplestreams.CloudSpec, error) {
	endpoint, err := gosigma.ResolveEndpoint(region)
	if err != nil {
		return simplestreams.CloudSpec{}, err
	}
	return simplestreams.CloudSpec{
		Region:   region,
		Endpoint: endpoint,
	}, nil
}
