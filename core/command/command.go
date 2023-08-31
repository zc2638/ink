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
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zc2638/ink/core/clients"
	"github.com/zc2638/ink/core/constant"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/files"
)

func Register(cmd *cobra.Command, name string, opts ...any) {
	subCmd := &cobra.Command{
		Use:          name,
		SilenceUsage: true,
	}
	for _, opt := range opts {
		switch v := opt.(type) {
		case func(*cobra.Command, []string):
			subCmd.Run = v
		case func(*cobra.Command, []string) error:
			subCmd.RunE = v
		case *flag.FlagSet:
			subCmd.Flags().AddGoFlagSet(v)
		case *flag.Flag:
			subCmd.Flags().AddGoFlag(v)
		case func(*cobra.Command):
			v(subCmd)
		}
	}
	cmd.AddCommand(subCmd)
}

func newServerClient(cmd *cobra.Command) (clients.ServerV1, error) {
	serverAddr, err := cmd.Flags().GetString("server")
	if err != nil {
		return nil, err
	}
	sc, err := clients.NewServer(serverAddr)
	if err != nil {
		return nil, fmt.Errorf("init client failed: %v", err)
	}
	return sc.V1(), nil
}

func getEnv(name string) string {
	return getDefaultEnv(name, "")
}

func getDefaultEnv(name string, defValue string) string {
	key := strings.ToUpper(constant.Name + "_" + name)
	value := os.Getenv(key)
	if value == "" {
		value = defValue
	}
	return value
}

func getNN(args []string) (namespace, name string, err error) {
	if len(args) == 0 {
		err = constant.ErrInvalidName
		return
	}

	nn := args[0]
	parts := strings.SplitN(nn, "/", 2)
	if len(parts) == 1 {
		namespace = v1.DefaultNamespace
		name = parts[0]
	} else {
		namespace = parts[0]
		name = parts[1]
	}
	return
}

func getPage(cmd *cobra.Command) *v1.Pagination {
	f := cmd.Flags()
	page, _ := f.GetInt("page")
	size, _ := f.GetInt("size")
	return &v1.Pagination{Page: page, Size: size}
}

func getFileData(cmd *cobra.Command) ([]files.Item, error) {
	fp, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}
	if len(fp) == 0 {
		return nil, errors.New("file path is not defined")
	}

	stat, err := os.Stat(fp)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		item, err := files.ReadFile(fp)
		if err != nil {
			return nil, err
		}
		return []files.Item{*item}, nil
	}

	dirs, err := os.ReadDir(fp)
	if err != nil {
		return nil, err
	}
	var set []files.Item
	for _, entry := range dirs {
		if entry.IsDir() {
			continue
		}

		name := filepath.Join(fp, entry.Name())
		item, err := files.ReadFile(name)
		if errors.Is(err, files.ErrUnknownExt) {
			continue
		}
		if err != nil {
			return nil, err
		}
		set = append(set, *item)
	}
	return set, nil
}

func write(b []byte) {
	fmt.Println(string(b))
}

func writeString(s string) {
	fmt.Println(s)
}
