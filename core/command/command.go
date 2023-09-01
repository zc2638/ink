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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zc2638/ink/core/clients"
	"github.com/zc2638/ink/core/constant"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/files"
	"github.com/zc2638/ink/pkg/utils"
)

func Register(cmd *cobra.Command, name string, short string, opts ...any) {
	subCmd := new(cobra.Command)
	for _, opt := range opts {
		v, ok := opt.(*cobra.Command)
		if !ok {
			continue
		}
		subCmd = v
	}

	subCmd.Use = name
	subCmd.SilenceUsage = true
	subCmd.Short = short

	for _, opt := range opts {
		switch v := opt.(type) {
		case *cobra.Command:
			continue
		case func(*cobra.Command, []string):
			subCmd.Run = v
		case func(*cobra.Command, []string) error:
			subCmd.RunE = v
		case *flag.FlagSet:
			subCmd.Flags().AddGoFlagSet(v)
		case *flag.Flag:
			subCmd.Flags().AddGoFlag(v)
		case Example:
			subCmd.Example = IndentLine(v.String())
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

func IndentLine(s string) string {
	return Indent(2, s)
}

func Indent(n int, s string) string {
	if n < 1 {
		return s
	}
	spaces := fmt.Sprintf("%"+strconv.Itoa(n)+"s", "")

	s = strings.TrimSpace(s)
	parts := strings.Split(s, "\n")
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(spaces)
		b.WriteString(part)
		b.WriteString("\n")
	}
	return b.String()
}

func parseObjects(cmd *cobra.Command) (map[string][]v1.UnstructuredObject, error) {
	items, err := getFileData(cmd)
	if err != nil {
		return nil, err
	}

	objSet := make(map[string][]v1.UnstructuredObject)
	for _, item := range items {
		jsonBytes := item.Data
		if !utils.IsJSON(item.Data) {
			jsonBytes, err = utils.ConvertYAMLToJSON(item.Data)
			if err != nil {
				return nil, err
			}
		}

		var objs []v1.UnstructuredObject
		if utils.IsJSONArray(jsonBytes) {
			if err := json.Unmarshal(jsonBytes, &objs); err != nil {
				return nil, err
			}
		} else {
			var obj v1.UnstructuredObject
			if err := json.Unmarshal(jsonBytes, &obj); err != nil {
				return nil, err
			}
			objs = []v1.UnstructuredObject{obj}
		}
		for _, obj := range objs {
			kind := obj.GetKind()
			objSet[kind] = append(objSet[kind], obj)
		}
	}
	return objSet, nil
}
