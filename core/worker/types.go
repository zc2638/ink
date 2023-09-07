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

package worker

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/zc2638/ink/core/constant"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

type State struct {
	ExitCode  int
	OOMKilled bool
}

type Workflow struct {
	ID        string
	Name      string
	Namespace string
	Labels    map[string]string

	Steps       []*Step
	WorkingDir  string
	Concurrency int
	Volumes     []Volume
	DependsOn   []string
	Worker      *v1.Worker
}

func (s *Workflow) GetStep(name string) *Step {
	for _, v := range s.Steps {
		if v.Name == name {
			return v
		}
	}
	return nil
}

type (
	Volume struct {
		v1.Volume

		ID string
	}

	VolumeDevice struct {
		Name string
		Path string
	}
)

type Step struct {
	ID              string
	Name            string
	Image           string
	ImagePullPolicy v1.PullPolicy
	ImagePullAuth   string
	Privileged      bool
	WorkingDir      string
	Network         string
	Env             map[string]string
	DNS             []string
	DNSSearch       []string
	ExtraHosts      []string
	Entrypoint      []string
	Shell           []string
	Command         []string
	Args            []string
	VolumeMounts    []v1.VolumeMount
	Devices         []v1.VolumeDevice
}

func (s *Step) CombineEnv(env ...any) map[string]string {
	out := maps.Clone(s.Env)
	for _, v := range env {
		switch value := v.(type) {
		case map[string]string:
			for key, val := range value {
				out[key] = val
			}
		case string:
			parts := strings.SplitN(value, "=", 2)
			if len(parts) == 2 {
				out[parts[0]] = parts[1]
			}
		}
	}
	return out
}

func completeID(id string) string {
	return constant.Name + "-" + id
}

func Convert(in *v1.Workflow, status *v1.Stage, secrets []*v1.Secret) (*Workflow, error) {
	out := &Workflow{
		ID:          completeID(strconv.FormatUint(status.ID, 10)),
		Name:        in.Name,
		Namespace:   in.Namespace,
		Labels:      in.Labels,
		WorkingDir:  in.Spec.WorkingDir,
		Concurrency: in.Spec.Concurrency,
		DependsOn:   in.Spec.DependsOn,
		Worker:      in.Spec.Worker,
	}

	for _, v := range in.Spec.Volumes {
		out.Volumes = append(out.Volumes, Volume{Volume: v})
	}

	imagePullSecrets := make([]*v1.Secret, 0)
	for _, v := range in.Spec.ImagePullSecrets {
		for _, sv := range secrets {
			if v == sv.Name {
				imagePullSecrets = append(imagePullSecrets, sv)
				break
			}
		}
	}

	for _, v := range in.Spec.Steps {
		var id string
		for _, vv := range status.Steps {
			if vv.Name == v.Name {
				id = strconv.FormatUint(vv.ID, 10)
				break
			}
		}
		if id == "" {
			return nil, fmt.Errorf("step not found: %s", v.Name)
		}

		step := &Step{
			ID:              completeID(id),
			Name:            v.Name,
			Image:           v.Image,
			ImagePullPolicy: v.ImagePullPolicy,
			Privileged:      v.Privileged,
			WorkingDir:      v.WorkingDir,
			Entrypoint:      v.Entrypoint,
			Shell:           v.Shell,
			Command:         v.Command,
			Args:            v.Args,
			VolumeMounts:    v.VolumeMounts,
			Devices:         v.Devices,
			DNS:             v.DNS,
			DNSSearch:       v.DNSSearch,
			ExtraHosts:      v.ExtraHosts,
		}

		// image registry auth
		for _, sv := range imagePullSecrets {
			_ = sv.Decrypt()
			dockerAuthsData, ok := sv.Data[v1.DockerConfigJSONKey]
			if !ok {
				continue
			}

			var dockerAuths v1.DockerAuths
			if err := json.Unmarshal([]byte(dockerAuthsData), &dockerAuths); err != nil {
				continue
			}
			step.ImagePullAuth = dockerAuths.Match(v.Image)
			break
		}

		env := make(map[string]string)
		for _, ev := range v.Env {
			if ev.Name == "" {
				continue
			}
			if ev.Value != "" {
				env[ev.Name] = ev.Value
				continue
			}
			// secret to env
			if ev.ValueFrom != nil && ev.ValueFrom.SecretKeyRef != nil {
				_, secData := ev.ValueFrom.SecretKeyRef.Find(secrets)
				env[ev.Name] = secData
			}
		}
		for _, sv := range v.Settings {
			if sv.Name == "" || sv.Value == "" {
				continue
			}
			env[sv.Name] = sv.Value
		}
		if len(env) > 0 {
			step.Env = env
		}
		out.Steps = append(out.Steps, step)
	}

	Compile(out)
	return out, nil
}

func Compile(spec *Workflow) {
	if len(spec.WorkingDir) == 0 {
		spec.WorkingDir = constant.WorkspacePath
	} else if !filepath.IsAbs(spec.WorkingDir) {
		spec.WorkingDir = filepath.Join(constant.WorkspacePath, spec.WorkingDir)
	}

	if IsRestrictedVolume(spec.WorkingDir) {
		spec.WorkingDir = constant.WorkspacePath
	}

	// add the workspace volume
	volume := Volume{
		ID: spec.ID,
		Volume: v1.Volume{
			Name:     "_ink_volume",
			EmptyDir: &v1.EmptyDirVolume{},
		},
	}
	spec.Volumes = append([]Volume{volume}, spec.Volumes...)

	for _, s := range spec.Steps {
		switch s.ImagePullPolicy {
		case v1.PullAlways, v1.PullNever, v1.PullIfNotPresent:
		default:
			s.ImagePullPolicy = v1.PullIfNotPresent
		}

		if s.WorkingDir == "" {
			s.WorkingDir = spec.WorkingDir
		}
		vm := v1.VolumeMount{
			Name: volume.Name,
			Path: s.WorkingDir,
		}
		s.VolumeMounts = append(s.VolumeMounts, vm)
	}
}

// IsRestrictedVolume is a helper function that
// returns true if mounting the volume is restricted for untrusted containers.
func IsRestrictedVolume(path string) bool {
	path, err := filepath.Abs(path)
	if err != nil {
		return true
	}

	path = strings.ToLower(path)

	switch {
	case path == "/":
	case path == "/etc":
	case path == "/etc/docker" || strings.HasPrefix(path, "/etc/docker/"):
	case path == "/var":
	case path == "/var/run" || strings.HasPrefix(path, "/var/run/"):
	case path == "/proc" || strings.HasPrefix(path, "/proc/"):
	case path == "/usr/local/bin" || strings.HasPrefix(path, "/usr/local/bin/"):
	case path == "/usr/local/sbin" || strings.HasPrefix(path, "/usr/local/sbin/"):
	case path == "/usr/bin" || strings.HasPrefix(path, "/usr/bin/"):
	case path == "/bin" || strings.HasPrefix(path, "/bin/"):
	case path == "/mnt" || strings.HasPrefix(path, "/mnt/"):
	case path == "/mount" || strings.HasPrefix(path, "/mount/"):
	case path == "/media" || strings.HasPrefix(path, "/media/"):
	case path == "/sys" || strings.HasPrefix(path, "/sys/"):
	case path == "/dev" || strings.HasPrefix(path, "/dev/"):
	default:
		return false
	}

	return true
}

func EnvToSlice(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for ek, ev := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", ek, ev))
	}
	sort.Strings(env)
	return env
}
