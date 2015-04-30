// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package upgrades

import "github.com/juju/juju/state"

// stateStepsFor124 returns upgrade steps for Juju 1.24 that manipulate state directly.
func stateStepsFor124() []Step {
	var steps []Step
	steps = append(steps,
		&upgradeStep{
			description: "add block device documents for existing machines",
			targets:     []Target{DatabaseMaster},
			run: func(context Context) error {
				return state.AddDefaultBlockDevicesDocs(context.State())
			},
		},
	)
	return steps
}
