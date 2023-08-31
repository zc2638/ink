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

package command

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zc2638/wslog"
	"golang.org/x/sync/errgroup"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/worker"
	"github.com/zc2638/ink/core/worker/hooks"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

func NewWorker() *cobra.Command {
	opt := new(WorkerOption)
	opt.ConfigPath = DefaultConfig(constant.WorkerName)

	cmd := &cobra.Command{
		Use:          constant.WorkerName,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg WorkerConfig
			_, err := ParseAllConfig(opt.ConfigPath, &cfg, constant.DaemonName, opt.ConfigSubKey)
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}
			wslog.Infof("Config: %#v", cfg)

			logger := wslog.New(cfg.Logger)

			workers := make([]*worker.Worker, 0, len(cfg.Workers))
			for k, v := range cfg.Workers {
				if v.Worker == nil {
					v.Worker = &v1.Worker{Kind: v1.WorkerKindDocker}
				}

				// TODO
				var hook worker.Hook
				switch v.Worker.Kind {
				case v1.WorkerKindHost:
					hook, err = hooks.NewHost()
					if err != nil {
						return fmt.Errorf("init host hook failed: %v", err)
					}
				case v1.WorkerKindSSH:
				case v1.WorkerKindDocker:
					hook, err = hooks.NewDocker("", "")
					if err != nil {
						return fmt.Errorf("init docker hook failed: %v", err)
					}
				case v1.WorkerKindKubernetes:
				default:
					return fmt.Errorf("unsupported kind: %s", v.Worker.Kind)
				}

				w, err := worker.New(v, hook, logger)
				if err != nil {
					return fmt.Errorf("create worker(%d) failed: %v", k, err)
				}
				workers = append(workers, w)
			}

			eg, ctx := errgroup.WithContext(context.Background())
			for _, w := range workers {
				wc := w
				eg.Go(func() error { return wc.Run(ctx) })
			}
			return eg.Wait()
		},
	}

	cmd.PersistentFlags().StringVarP(&opt.ConfigPath, "config", "c", opt.ConfigPath, "config path")
	cmd.PersistentFlags().StringVar(&opt.ConfigSubKey, "config-sub-key", opt.ConfigSubKey, "config sub key for config data")
	return cmd
}

type WorkerOption struct {
	ConfigPath   string
	ConfigSubKey string
}

type WorkerConfig struct {
	Logger  wslog.Config    `json:"logger,omitempty"`
	Workers []worker.Config `json:"workers,omitempty"`
}

func (c *WorkerConfig) Validate() error {
	if len(c.Workers) == 0 {
		return errors.New("at least one worker is defined")
	}
	hostname, _ := os.Hostname()
	if len(hostname) == 0 {
		hostname = constant.WorkerName
	}
	for i, w := range c.Workers {
		// TODO auto name
		if len(w.Name) == 0 {
			c.Workers[i].Name = fmt.Sprintf("%s.%s", hostname, w.Worker.Kind)
		}
	}
	return nil
}
