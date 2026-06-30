// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"sync"
	"time"
)

var (
	DefaultSyncTime            = "5m"
	DefaultSyncProcessingDelay = "200ms"
)

// TargetNameT defines TODO
type TargetSelectorT struct {
	MatchExact string `yaml:"matchExact,omitempty"`
	MatchRegex string `yaml:"matchRegex,omitempty"`
}

// TargetT defines TODO
type TargetT struct {
	Group     string          `yaml:"group"`
	Version   string          `yaml:"version"`
	Resource  string          `yaml:"resource"`
	Name      TargetSelectorT `yaml:"name"`
	Namespace TargetSelectorT `yaml:"namespace"`
}

// ConditionT defines TODO
type ConditionT struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// ResourceT defines TODO
type ResourceT struct {
	Target     TargetT      `yaml:"target"`
	PreStep    string       `yaml:"preStep,omitempty"`
	Conditions []ConditionT `yaml:"conditions"`
}

// MetadataSpec TODO
type MetadataT struct {
	Name string `yaml:"name"`
}

// SynchronizationT defines TODO
type SynchronizationT struct {
	Time            string `yaml:"time"`
	ProcessingDelay string `yaml:"processingDelay,omitempty"`

	// Carried stuff
	CarriedTime            time.Duration
	CarriedProcessingDelay time.Duration
}

// SpecificationSpec TODO
type SpecificationT struct {
	Synchronization SynchronizationT `yaml:"synchronization"`
	Resources       []ResourceT      `yaml:"resources"`
}

// ConfigSpec TODO
type ConfigT struct {
	Mutex sync.RWMutex

	//
	ApiVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   MetadataT      `yaml:"metadata"`
	Spec       SpecificationT `yaml:"spec"`
}
