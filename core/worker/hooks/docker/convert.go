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

package docker

import (
	"os"
	"strings"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"

	"github.com/zc2638/ink/core/worker"
	"github.com/zc2638/ink/pkg/shell"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

func toContainerConfig(stage *worker.Stage, step *worker.Step) *container.Config {
	env := step.CombineEnv(map[string]string{
		"INK_SCRIPT": shell.Script(step.Command),
	})

	cfg := &container.Config{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		OpenStdin:    false,
		StdinOnce:    false,
		ArgsEscaped:  false,
		WorkingDir:   step.WorkingDir,
		Image:        step.Image,
		Env:          worker.EnvToSlice(env),
	}

	if len(step.Entrypoint) > 0 {
		cfg.Entrypoint = step.Entrypoint
		cfg.Cmd = step.Command
	} else {
		cmdName, args := shell.Command()
		if len(step.Shell) > 0 {
			cmdName = step.Shell[0]
			args = step.Shell[1:]
		}
		cfg.Entrypoint = append([]string{cmdName}, args...)
		cfg.Cmd = shell.EchoEnvCommand("INK_SCRIPT", cmdName)
	}

	for _, sec := range step.Secrets {
		cfg.Env = append(cfg.Env, sec.Name+"="+sec.Data)
	}
	if len(step.VolumeMounts) != 0 {
		cfg.Volumes = toVolumeSet(stage, step)
	}
	return cfg
}

func toHostConfig(stage *worker.Stage, step *worker.Step) *container.HostConfig {
	config := &container.HostConfig{
		Privileged: step.Privileged,
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
	}
	// windows do not support privileged, so we hard-code
	// this value to false.
	if stage.Worker != nil && stage.Worker.Platform.OS == "windows" {
		config.Privileged = false
	}
	if len(step.Network) > 0 {
		config.NetworkMode = container.NetworkMode(step.Network)
	}
	if len(step.DNS) > 0 {
		config.DNS = step.DNS
	}
	if len(step.DNSSearch) > 0 {
		config.DNSSearch = step.DNSSearch
	}
	if len(step.ExtraHosts) > 0 {
		config.ExtraHosts = step.ExtraHosts
	}
	if !isUnlimited(step) {
		// TODO
		config.Resources = container.Resources{}
	}

	if len(step.VolumeMounts) > 0 {
		config.Devices = toDeviceSlice(stage, step)
		config.Binds = toVolumeSlice(stage, step)
		config.Mounts = toVolumeMounts(stage, step)
	}
	return config
}

func toNetConfig(stage *worker.Stage, step *worker.Step) *network.NetworkingConfig {
	// if the user overrides the default network,
	// we do not attach to the user-defined network.
	if step.Network != "" {
		return &network.NetworkingConfig{}
	}
	endpoints := map[string]*network.EndpointSettings{}
	endpoints[stage.ID] = &network.EndpointSettings{
		NetworkID: stage.ID,
		Aliases:   []string{step.Name},
	}
	return &network.NetworkingConfig{
		EndpointsConfig: endpoints,
	}
}

// toDeviceSlice converts a slice of device paths to a slice of container.DeviceMapping.
func toDeviceSlice(stage *worker.Stage, step *worker.Step) []container.DeviceMapping {
	var to []container.DeviceMapping
	for _, vm := range step.Devices {
		device, ok := lookupVolume(stage, vm.Name)
		if !ok {
			continue
		}
		if !isDevice(device) {
			continue
		}
		to = append(to, container.DeviceMapping{
			PathOnHost:        device.HostPath.Path,
			PathInContainer:   vm.Path,
			CgroupPermissions: "rwm",
		})
	}
	if len(to) == 0 {
		return nil
	}
	return to
}

// helper function that converts a slice of volume paths to a set
// of unique volume names.
func toVolumeSet(stage *worker.Stage, step *worker.Step) map[string]struct{} {
	set := map[string]struct{}{}
	for _, vm := range step.VolumeMounts {
		volume, ok := lookupVolume(stage, vm.Name)
		if !ok {
			continue
		}
		if isDevice(volume) {
			continue
		}
		if isNamedPipe(volume) {
			continue
		}
		if !isBindMount(volume) {
			continue
		}
		set[vm.Path] = struct{}{}
	}
	return set
}

// toVolumeSlice returns a slice of volume mounts.
func toVolumeSlice(stage *worker.Stage, step *worker.Step) []string {
	// this entire function should be deprecated in favor of toVolumeMounts.
	// however, I am unable to get it working with data volumes.
	var to []string
	for _, vm := range step.VolumeMounts {
		volume, ok := lookupVolume(stage, vm.Name)
		if !ok {
			continue
		}
		if isDevice(volume) {
			continue
		}
		if isDataVolume(volume) {
			path := volume.ID + ":" + vm.Path
			to = append(to, path)
		}
		if isBindMount(volume) {
			path := volume.HostPath.Path + ":" + vm.Path
			to = append(to, path)
		}
	}
	return to
}

// toVolumeMounts returns a slice of docker mount configurations.
func toVolumeMounts(stage *worker.Stage, step *worker.Step) []mount.Mount {
	var mounts []mount.Mount
	for _, vm := range step.VolumeMounts {
		source, ok := lookupVolume(stage, vm.Name)
		if !ok {
			continue
		}
		if isBindMount(source) && !isDevice(source) {
			continue
		}
		if isDataVolume(source) {
			continue
		}
		mounts = append(mounts, toMount(source, &vm))
	}
	if len(mounts) == 0 {
		return nil
	}
	return mounts
}

// helper function converts the volume declaration to a
// docker mount structure.
func toMount(source *worker.Volume, target *v1.VolumeMount) mount.Mount {
	to := mount.Mount{
		Target: target.Path,
		Type:   toVolumeType(source),
	}
	if isBindMount(source) || isNamedPipe(source) {
		to.Source = source.HostPath.Path
	}
	if isTmpfs(source) {
		to.TmpfsOptions = &mount.TmpfsOptions{
			SizeBytes: int64(source.EmptyDir.SizeLimit),
			Mode:      os.ModePerm,
		}
	}
	return to
}

// toVolumeType returns the docker volume enumeration
// for the given volume.
func toVolumeType(from *worker.Volume) mount.Type {
	switch {
	case isDataVolume(from):
		return mount.TypeVolume
	case isTmpfs(from):
		return mount.TypeTmpfs
	case isNamedPipe(from):
		return mount.TypeNamedPipe
	default:
		return mount.TypeBind
	}
}

// isUnlimited returns true if the container has no resource limits.
func isUnlimited(_ *worker.Step) bool {
	// TODO
	return true
}

// isBindMount returns true if the volume is a bind mount.
func isBindMount(volume *worker.Volume) bool {
	return volume.HostPath != nil
}

// isTmpfs returns true if the volume is in-memory.
func isTmpfs(volume *worker.Volume) bool {
	return volume.EmptyDir != nil && volume.EmptyDir.Medium == "memory"
}

// isDataVolume returns true if the volume is a data-volume.
func isDataVolume(volume *worker.Volume) bool {
	return volume.EmptyDir != nil && volume.EmptyDir.Medium != "memory"
}

// isDevice returns true if the volume is a device
func isDevice(volume *worker.Volume) bool {
	return volume.HostPath != nil && strings.HasPrefix(volume.HostPath.Path, "/dev/")
}

// isNamedPipe returns true if the volume is a named pipe.
func isNamedPipe(volume *worker.Volume) bool {
	return volume.HostPath != nil &&
		strings.HasPrefix(volume.HostPath.Path, `\\.\pipe\`)
}

// lookupVolume returns the named volume.
func lookupVolume(stage *worker.Stage, name string) (*worker.Volume, bool) {
	for _, v := range stage.Volumes {
		if v.HostPath != nil && v.Name == name {
			return &v, true
		}
		if v.EmptyDir != nil && v.Name == name {
			return &v, true
		}
	}
	return nil, false
}
