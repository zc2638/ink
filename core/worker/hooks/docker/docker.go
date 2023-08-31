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
	"context"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/zc2638/wslog"

	"github.com/zc2638/ink/core/worker"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

func New(host, version string) (worker.Hook, error) {
	opts := []client.Opt{client.FromEnv}
	if len(host) > 0 {
		opts = append(opts, client.WithHost(host))
	}
	if len(version) > 0 {
		opts = append(opts, client.WithVersion(version))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return &docker{client: cli}, nil
}

type docker struct {
	client client.APIClient
	states sync.Map
}

func (h *docker) getState(id string) *dockerState {
	v, ok := h.states.Load(id)
	if !ok {
		return nil
	}
	return v.(*dockerState)
}

func (h *docker) addState(id string, state *dockerState) {
	h.states.Store(id, state)
}

func (h *docker) delState(id string) {
	h.states.Delete(id)
}

func (h *docker) StageBegin(ctx context.Context, stage *worker.Stage) error {
	state := newDockerState()
	h.addState(stage.ID, state)

	state.id = stage.ID
	for _, v := range stage.Volumes {
		if v.EmptyDir == nil {
			continue
		}

		state.volumes[v.Name] = v.ID
		_, err := h.client.VolumeCreate(ctx, volume.CreateOptions{
			Name:   v.ID,
			Driver: "local",
		})
		if err != nil {
			return trimExtraInfo(err)
		}
	}

	driver := "bridge"
	if stage.Worker != nil &&
		stage.Worker.Platform != nil &&
		stage.Worker.Platform.OS == "windows" {
		driver = "nat"
	}
	_, err := h.client.NetworkCreate(ctx, stage.ID, types.NetworkCreate{Driver: driver})
	return trimExtraInfo(err)
}

func (h *docker) StageEnd(ctx context.Context, stage *worker.Stage) error {
	state := h.getState(stage.ID)
	if state == nil {
		return errors.New("abnormal state")
	}

	removeOpts := types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: true,
	}
	log := wslog.FromContext(ctx)

	// stop step containers
	for _, id := range state.containers {
		if err := h.client.ContainerKill(ctx, id, "9"); err != nil && !client.IsErrNotFound(err) && !errdefs.IsConflict(err) {
			log.Error("Kill container failed",
				"error", err,
				"container", id,
			)
		}
	}
	// remove step containers
	for _, id := range state.containers {
		if err := h.client.ContainerRemove(ctx, id, removeOpts); err != nil && !client.IsErrNotFound(err) {
			log.Error("Remove container failed",
				"error", err,
				"container", id,
			)
		}
	}

	for _, v := range state.volumes {
		if err := h.client.VolumeRemove(ctx, v, true); err != nil {
			log.Error("Remove volume failed",
				"error", err,
				"volumeID", v,
			)
		}
	}
	if err := h.client.NetworkRemove(ctx, state.id); err != nil {
		log.Error("Remove network failed",
			"error", err,
			"network", state.id,
		)
	}
	h.delState(stage.ID)
	return nil
}

func (h *docker) Step(ctx context.Context, stage *worker.Stage, step *worker.Step, writer io.Writer) (*worker.State, error) {
	log := wslog.FromContext(ctx).With("stage", stage.Name, "step", step.Name)

	state := h.getState(stage.ID)
	if state == nil {
		return nil, errors.New("abnormal state")
	}
	state.containers[step.Name] = step.ID

	image := ImageExpand(step.Image)
	isLatest := strings.HasSuffix(image, ":latest")
	pullOpts := types.ImagePullOptions{}

	// TODO docker registry auth
	//authMap := map[string]string{
	//	"username": "",
	//	"password": "",
	//}
	//buf, _ := json.Marshal(&authMap)
	//authStr := base64.URLEncoding.EncodeToString(buf)
	//pullopts.RegistryAuth = authStr

	if step.ImagePullPolicy == v1.PullIfNotPresent {
		var imageExist bool
		if !isLatest {
			searchImage := strings.TrimPrefix(image, "docker.io/library/")
			imageList, err := h.client.ImageList(ctx, types.ImageListOptions{
				Filters: filters.NewArgs(filters.Arg("reference", searchImage)),
			})
			if err != nil {
				return nil, err
			}
			imageExist = len(imageList) > 0
		}
		if !imageExist {
			rc, pullErr := h.client.ImagePull(ctx, image, pullOpts)
			if pullErr != nil {
				return nil, pullErr
			}
			_ = PullReaderCopy(rc, writer)
			rc.Close()
		}
	} else if step.ImagePullPolicy == v1.PullAlways {
		rc, pullErr := h.client.ImagePull(ctx, image, pullOpts)
		if pullErr != nil {
			return nil, pullErr
		}
		_ = PullReaderCopy(rc, writer)
		rc.Close()
	}

	containerConfig := toContainerConfig(stage, step)
	hostConfig := toHostConfig(stage, step)
	networkConfig := toNetConfig(stage, step)
	_, err := h.client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, step.ID)
	if err != nil {
		return nil, err
	}

	// TODO user defined network connect
	if err := h.client.ContainerStart(ctx, step.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}
	logs, err := h.client.ContainerLogs(ctx, step.ID, types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	})
	if err != nil {
		return nil, err
	}
	_, _ = StdCopy(writer, logs)
	defer logs.Close()

	waitCh, errCh := h.client.ContainerWait(ctx, step.ID, container.WaitConditionNotRunning)
	select {
	case err = <-errCh:
		if errdefs.IsCancelled(err) {
			return nil, context.Canceled
		}
		log.With("error", err).Error("container wait error")
	case wait := <-waitCh:
		log.With(
			"statusCode", wait.StatusCode,
			"container", step.ID,
		).Debug("container wait response")
	}

	info, err := h.client.ContainerInspect(ctx, step.ID)
	if err != nil {
		return nil, err
	}

	return &worker.State{
		ExitCode:  info.State.ExitCode,
		OOMKilled: info.State.OOMKilled,
	}, nil
}

// trimExtraInfo is a helper function that trims extra information
// from a Docker error. Specifically, on Windows, this can expose
// environment variables and other sensitive data.
func trimExtraInfo(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	i := strings.Index(s, "extra info:")
	if i > 0 {
		s = s[:i]
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "(0x2)")
		s = strings.TrimSpace(s)
		return errors.New(s)
	}
	return err
}

// ImageExpand returns the fully qualified image name.
func ImageExpand(name string) string {
	ref, err := reference.ParseAnyReference(name)
	if err != nil {
		return name
	}
	named, err := reference.ParseNamed(ref.String())
	if err != nil {
		return name
	}
	named = reference.TagNameOnly(named)
	return named.String()
}
