// Copyright © 2023 zc2638 <zc2638@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

type Workflow struct {
	Metadata `yaml:",inline"`

	Spec WorkflowSpec `json:"spec" yaml:"spec"`
}

func (w *Workflow) Worker() *Worker {
	if w.Spec.Worker != nil {
		return w.Spec.Worker
	}
	return &Worker{
		Kind: WorkerKindDocker,
	}
}

type WorkflowSpec struct {
	Steps       []Flow   `json:"steps" yaml:"steps"`
	WorkingDir  string   `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Concurrency int      `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	Volumes     []Volume `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	DependsOn   []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Worker      *Worker  `json:"worker,omitempty" yaml:"worker,omitempty"`
}

type Flow struct {
	Name            string         `json:"name" yaml:"name"`
	Image           string         `json:"image,omitempty" yaml:"image,omitempty"`
	ImagePullPolicy PullPolicy     `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Privileged      bool           `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	WorkingDir      string         `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Env             []EnvVar       `json:"env,omitempty" yaml:"env,omitempty"`
	Entrypoint      []string       `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Shell           []string       `json:"shell,omitempty" yaml:"shell,omitempty"`
	Command         []string       `json:"command,omitempty" yaml:"command,omitempty"`
	Args            []string       `json:"args,omitempty" yaml:"args,omitempty"`
	VolumeMounts    []VolumeMount  `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	Devices         []VolumeDevice `json:"devices,omitempty" yaml:"devices,omitempty"`
	DNS             []string       `json:"dns,omitempty" yaml:"dns,omitempty"`
	DNSSearch       []string       `json:"dnsSearch,omitempty" yaml:"dnsSearch,omitempty"`
	ExtraHosts      []string       `json:"extraHosts,omitempty" yaml:"extraHosts,omitempty"`
}

type PullPolicy string

func (s PullPolicy) String() string { return string(s) }

const (
	// PullAlways means that kubelet always attempts to pull the latest image.
	// Flow will fail If the pull fails.
	PullAlways PullPolicy = "Always"
	// PullNever means that kubelet never pulls an image, but only uses a local image.
	// Flow will fail if the image isn't present.
	PullNever PullPolicy = "Never"
	// PullIfNotPresent means that kubelet pulls if the image isn't present on disk.
	// Flow will fail if the image isn't present and the pull fails.
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

type Volume struct {
	Name string `json:"name"`

	HostPath *HostPathVolume `json:"hostPath,omitempty" `
	EmptyDir *EmptyDirVolume `json:"emptyDir,omitempty" `
}

type HostPathVolume struct {
	Path string `json:"path"`
}

// VolumeMount describes a mounting of a Volume within a container.
type VolumeMount struct {
	// This must match the Name of a Volume.
	Name string `json:"name"`
	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	Path string `json:"path"`
}

type EmptyDirVolume struct {
	Medium    StorageMedium `json:"medium,omitempty"`
	SizeLimit BytesSize     `json:"sizeLimit,omitempty"`
}

// StorageMedium defines ways that storage can be allocated to a volume.
type StorageMedium string

const (
	StorageMediumDefault StorageMedium = ""       // use whatever the default is for the node, assume anything we don't explicitly handle is this
	StorageMediumMemory  StorageMedium = "memory" // use memory (e.g. tmpfs on linux)
)

type VolumeDevice struct {
	Name string `json:"name" yaml:"name"`
	Path string `json:"path" yaml:"path"`
}

// EnvVar represents an environment variable present in a Flow.
type EnvVar struct {
	// Name of the environment variable.
	Name string `json:"name" yaml:"name"`

	// Variable references $(VAR_NAME) are expanded
	// using the previously defined environment variables in the container and
	// any service environment variables. If a variable cannot be resolved,
	// the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
	// "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
	// Escaped references will never be expanded, regardless of whether the variable exists or not.
	// Defaults to "".
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	// Source for the environment variable's value. Cannot be used if the value is not empty.
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for the value of an EnvVar.
type EnvVarSource struct {
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty" protobuf:"bytes,4,opt,name=secretKeyRef"`
}

// SecretKeySelector selects a key of a Secret.
type SecretKeySelector struct {
	// The name of the secret in the same namespace to select from.
	Name string `json:"name" yaml:"name"`
	// The key of the secret to select from.  Must be a valid secret key.
	Key string `json:"key" yaml:"key"`
}

func (s *SecretKeySelector) Find(secrets []*Secret) string {
	for _, v := range secrets {
		if v.Name != s.Name {
			continue
		}
		_ = v.Decrypt()
		return v.Data[s.Key]
	}
	return ""
}
