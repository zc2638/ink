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

package host

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zc2638/wslog"

	"github.com/zc2638/ink/core/worker"
	"github.com/zc2638/ink/pkg/shell"
)

func New() (worker.Hook, error) {
	return &host{}, nil
}

type host struct{}

func (h *host) Begin(ctx context.Context, spec *worker.Workflow) error {
	log := wslog.FromContext(ctx)

	homedir := getHomedir(spec)
	scriptPath := filepath.Join(homedir, "scripts")
	if err := os.MkdirAll(scriptPath, os.ModePerm); err != nil {
		return err
	}

	for _, step := range spec.Steps {
		cmdName, args := shell.Command()
		if len(step.Shell) > 0 {
			cmdName = step.Shell[0]
			args = step.Shell[1:]
		}
		cmdData := shell.Script(step.Command)
		fp := filepath.Join(scriptPath, step.Name+shell.Suffix)
		if err := os.WriteFile(fp, []byte(cmdData), os.ModePerm); err != nil {
			log.Error("cannot write file", "error", err)
			return err
		}
		args = append(args, fp)
		step.Args = args
		step.Command = []string{cmdName}
	}
	return nil
}

func (h *host) End(_ context.Context, spec *worker.Workflow) error {
	return os.RemoveAll(getRootDir(spec))
}

func (h *host) Step(ctx context.Context, spec *worker.Workflow, step *worker.Step, writer io.Writer) (*worker.State, error) {
	if len(step.Command) == 0 {
		return nil, nil
	}

	homedir := getHomedir(spec)
	rootDir := getRootDir(spec)
	workingDir := filepath.Join(rootDir, step.WorkingDir)
	if err := os.MkdirAll(workingDir, os.ModePerm); err != nil {
		return nil, err
	}

	env := step.CombineEnv(map[string]string{
		"HOME":          homedir,
		"HOMEPATH":      homedir, // for windows
		"USERPROFILE":   homedir, // for windows
		"INK_HOME":      workingDir,
		"INK_WORKSPACE": workingDir,
	})

	cmd := exec.CommandContext(ctx, step.Command[0], step.Args...)
	cmd.Env = worker.EnvToSlice(env)
	cmd.Dir = workingDir
	cmd.Stdout = writer
	cmd.Stderr = writer

	//for _, secret := range step.Secrets {
	//	s := fmt.Sprintf("%s=%s", secret.Env, string(secret.Data))
	//	cmd.Env = append(cmd.Env, s)
	//}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	log := wslog.FromContext(ctx)
	log = log.With("process.pid", cmd.Process.Pid)
	log.Debug("process started")

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err = <-done:
	case <-ctx.Done():
		_ = cmd.Process.Kill()

		log.Debug("process killed")
		return nil, ctx.Err()
	}

	state := new(worker.State)
	if err != nil {
		state.ExitCode = 255
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		state.ExitCode = exiterr.ExitCode()
	}

	log.Debug("process finished", "process.exit", state.ExitCode)
	return state, nil
}
