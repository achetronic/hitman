/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

// TargetT defines TODO
type TargetT struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ConditionT defines TODO
type ConditionT struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ResourceT defines TODO
type ResourceT struct {
	Target     TargetT      `json:"target"`
	Conditions []ConditionT `json:"conditions"`
}

// MetadataSpec TODO
type MetadataT struct {
	Name string `yaml:"name"`
}

// SynchronizationT defines TODO
type SynchronizationT struct {
	Time string `json:"time"`
}

// SpecificationSpec TODO
type SpecificationT struct {
	Synchronization SynchronizationT `yaml:"synchronization"`
	Resources       []ResourceT      `yaml:"resources"`
}

// ConfigSpec TODO
type ConfigT struct {
	ApiVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   MetadataT      `yaml:"metadata"`
	Spec       SpecificationT `yaml:"spec"`
}
