// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package client

var (
	RemoteParamsForMachine = remoteParamsForMachine
	GetAllUnitNames        = getAllUnitNames
	NewStateStorage        = &newStateStorage
)

// Filtering exports
var (
	MatchPortRanges = matchPortRanges
	MatchSubnet     = matchSubnet
)

// Status exports
var (
	ProcessMachines   = processMachines
	MakeMachineStatus = makeMachineStatus
)

type MachineAndContainers machineAndContainers
