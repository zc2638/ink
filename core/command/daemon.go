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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/99nil/gopkg/ctr"
	"github.com/99nil/gopkg/server"
	"github.com/99nil/gopkg/sets"
	"github.com/spf13/cobra"
	"github.com/zc2638/wslog"
	"gorm.io/gorm"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/handler"
	"github.com/zc2638/ink/core/scheduler"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
	"github.com/zc2638/ink/pkg/livelog"
	"github.com/zc2638/ink/pkg/queue"
	"github.com/zc2638/ink/resource"
)

func NewDaemon() *cobra.Command {
	opt := new(DaemonOption)
	opt.ConfigPath = DefaultConfig(constant.DaemonName)

	cmd := &cobra.Command{
		Use:          constant.DaemonName,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg DaemonConfig
			if _, err := ParseConfig(opt.ConfigPath, &cfg, constant.DaemonName, opt.ConfigSubKey); err != nil {
				if _, ok := err.(*os.PathError); !ok {
					return err
				}
				wslog.Warn("Config file not found, use default.")
			}
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("validate config failed: %v", err)
			}
			wslog.Infof("Config: %#v", cfg)

			log := wslog.New(cfg.Logger)
			ctr.SetLog(ctr.CoverKVLog(log))

			if err := database.AutoDatabase(cfg.Database); err != nil {
				return fmt.Errorf("auto database failed: %v", err)
			}

			db, err := database.New(cfg.Database)
			if err != nil {
				return fmt.Errorf("init database failed: %v", err)
			}
			if err := resource.MigrateDatabase(cfg.Database.Driver, cfg.Database.DSN); err != nil {
				return fmt.Errorf("migrate database failed: %v", err)
			}

			ll, err := livelog.New(cfg.Livelog)
			if err != nil {
				return fmt.Errorf("init livelog failed: %v", err)
			}
			sched := scheduler.New(listInCompleteStages(db))

			srv := server.New(&cfg.Server)
			srv.ReadTimeout = 0
			srv.WriteTimeout = 0
			srv.Handler = handler.New(log, db, ll, sched)
			log.Info(fmt.Sprintf("Daemon listen on %s", srv.Addr))
			return srv.RunAndStop(context.Background())
		},
	}

	cmd.PersistentFlags().StringVarP(&opt.ConfigPath, "config", "c", opt.ConfigPath, "config path")
	cmd.PersistentFlags().StringVar(&opt.ConfigSubKey, "config-sub-key", opt.ConfigSubKey, "config sub key for config data")
	return cmd
}

type DaemonOption struct {
	ConfigPath   string
	ConfigSubKey string
}

type DaemonConfig struct {
	Server   server.Config   `json:"server"`
	Logger   wslog.Config    `json:"logger,omitempty"`
	Database database.Config `json:"database,omitempty"`
	Queue    queue.Config    `json:"queue,omitempty"`
	Livelog  livelog.Config  `json:"livelog"`
}

func (c *DaemonConfig) Validate() error {
	if c.Server.Port <= 0 {
		c.Server.Port = 2638
	}
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite3"
		cacheDir := filepath.Join(os.TempDir(), constant.Name)
		c.Database.DSN = filepath.Join(cacheDir, "ink.db")
		if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
			return err
		}
	}
	if c.Livelog.File == nil {
		c.Livelog.File = &livelog.ConfigFile{
			Dir: filepath.Join(os.TempDir(), constant.Name, "cache"),
		}
	}
	return nil
}

func listInCompleteStages(db *gorm.DB) scheduler.StoreFunc {
	return func(ctx context.Context) ([]*v1.Stage, error) {
		db = db.WithContext(ctx)

		var list []storageV1.Stage
		if err := db.Where(&storageV1.Stage{Phase: v1.PhasePending.String()}).Find(&list).Error; err != nil {
			return nil, err
		}

		ids := sets.New[uint64]()
		for _, v := range list {
			ids.Add(v.BoxID)
		}
		if ids.Len() > 0 {
			var boxes []storageV1.Box
			if err := db.Where("id in (?)", ids.List()).Find(&boxes).Error; err != nil {
				return nil, err
			}
			for _, v := range boxes {
				ids.Remove(v.ID)
			}
			if ids.Len() > 0 {
				if err := db.Where("box_id in (?)", ids.List()).Updates(&storageV1.Stage{
					Phase:   v1.PhaseSkipped.String(),
					Started: time.Now().Unix(),
					Stopped: time.Now().Unix(),
				}).Error; err != nil {
					return nil, err
				}
			}
		}

		result := make([]*v1.Stage, 0, len(list))
		for _, v := range list {
			if ids.Has(v.BoxID) {
				continue
			}

			item, err := v.ToAPI()
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}
		return result, nil
	}
}
