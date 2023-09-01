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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/zc2638/ink/core/constant"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/flags"
	"github.com/zc2638/ink/pkg/printer"
	"github.com/zc2638/ink/pkg/utils"
)

func NewCtl() *cobra.Command {
	cmd := &cobra.Command{
		Use: constant.CtlName,
	}
	cmd.PersistentFlags().AddGoFlag(
		flags.New("server", flags.NewStringValue(getDefaultEnv("server", "http://localhost:2678")), ""),
	)
	Register(cmd, "apply", apply,
		flags.New("file", flags.NewStringValue(getEnv("file")), "that contains the configuration to apply"),
		func(cmd *cobra.Command) { cmd.Short = "apply a configuration to a resource by file name" },
	)

	workflowCmd := &cobra.Command{Use: "workflow", Short: "workflow operation"}
	Register(workflowCmd, "get", workflowGet, func(cmd *cobra.Command) { cmd.Short = "get workflow info" })
	Register(workflowCmd, "list", workflowList, func(cmd *cobra.Command) { cmd.Short = "list workflows" })
	Register(workflowCmd, "delete", workflowDelete, func(cmd *cobra.Command) { cmd.Short = "delete workflow" })

	boxCmd := &cobra.Command{Use: "box", Short: "box operation"}
	Register(boxCmd, "get", boxGet, func(cmd *cobra.Command) { cmd.Short = "get box info" })
	Register(boxCmd, "list", boxList, func(cmd *cobra.Command) { cmd.Short = "list boxes" })
	Register(boxCmd, "delete", boxDelete, func(cmd *cobra.Command) { cmd.Short = "delete box" })
	Register(boxCmd, "trigger", boxTrigger, func(cmd *cobra.Command) { cmd.Short = "create a build for box" })

	buildCmd := &cobra.Command{Use: "build", Short: "build operation"}
	Register(buildCmd, "get", buildGet, func(cmd *cobra.Command) { cmd.Short = "get build info" })
	Register(buildCmd, "list", buildList, func(cmd *cobra.Command) { cmd.Short = "list builds" })
	Register(buildCmd, "cancel", buildCancel, func(cmd *cobra.Command) { cmd.Short = "cancel build" })

	cmd.AddCommand(workflowCmd, boxCmd, buildCmd)
	return cmd
}

func apply(cmd *cobra.Command, _ []string) error {
	items, err := getFileData(cmd)
	if err != nil {
		return err
	}

	objSet := make(map[string][]v1.UnstructuredObject)
	for _, item := range items {
		jsonBytes := item.Data
		if !utils.IsJSON(item.Data) {
			jsonBytes, err = utils.ConvertYAMLToJSON(item.Data)
			if err != nil {
				return err
			}
		}

		var objs []v1.UnstructuredObject
		if utils.IsJSONArray(jsonBytes) {
			if err := json.Unmarshal(jsonBytes, &objs); err != nil {
				return err
			}
		} else {
			var obj v1.UnstructuredObject
			if err := json.Unmarshal(jsonBytes, &obj); err != nil {
				return err
			}
			objs = []v1.UnstructuredObject{obj}
		}
		for _, obj := range objs {
			kind := obj.GetKind()
			objSet[kind] = append(objSet[kind], obj)
		}
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// TODO kind Secret
	for _, obj := range objSet[v1.KindWorkflow] {
		var data v1.Workflow
		if err := obj.ToObject(&data); err != nil {
			return err
		}
		_, err := sc.WorkflowInfo(ctx, data.GetNamespace(), data.GetName())
		if err == nil {
			if err = sc.WorkflowUpdate(ctx, &data); err == nil {
				writeString(fmt.Sprintf("Update: %s", data.Metadata.String()))
			}
		} else if errors.Is(err, constant.ErrNoRecord) {
			if err = sc.WorkflowCreate(ctx, &data); err == nil {
				writeString(fmt.Sprintf("Create: %s", data.Metadata.String()))
			}
		}
		if err != nil {
			return err
		}
	}

	for _, obj := range objSet[v1.KindBox] {
		var data v1.Box
		if err := obj.ToObject(&data); err != nil {
			return err
		}
		_, err := sc.BoxInfo(ctx, data.GetNamespace(), data.GetName())
		if err == nil {
			if err = sc.BoxUpdate(ctx, &data); err == nil {
				writeString(fmt.Sprintf("Update: %s", data.Metadata.String()))
			}
		} else if errors.Is(err, constant.ErrNoRecord) {
			if err = sc.BoxCreate(ctx, &data); err == nil {
				writeString(fmt.Sprintf("Create: %s", data.Metadata.String()))
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func workflowGet(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, err := sc.WorkflowInfo(context.Background(), namespace, name)
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(result)
	if err != nil {
		return err
	}
	write(b)
	return nil
}

func workflowList(cmd *cobra.Command, _ []string) error {
	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, _, err := sc.WorkflowList(context.Background(), *getPage(cmd))
	if err != nil {
		return err
	}

	if len(result) == 0 {
		writeString("No resources found.")
		return nil
	}

	t := printer.NewTab("NAMESPACE", "NAME", "AGE")
	for _, v := range result {
		since := time.Since(v.Creation).Round(time.Second)
		t.Add(v.GetNamespace(), v.GetName(), since.String())
	}
	t.Print()
	return nil
}

func workflowDelete(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	return sc.WorkflowDelete(context.Background(), namespace, name)
}

func boxGet(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, err := sc.BoxInfo(context.Background(), namespace, name)
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(result)
	if err != nil {
		return err
	}
	write(b)
	return nil
}

func boxList(cmd *cobra.Command, _ []string) error {
	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, _, err := sc.BoxList(context.Background(), *getPage(cmd))
	if err != nil {
		return err
	}

	if len(result) == 0 {
		writeString("No resources found.")
		return nil
	}

	t := printer.NewTab("NAMESPACE", "NAME", "AGE")
	for _, v := range result {
		since := time.Since(v.Creation).Round(time.Second)
		t.Add(v.GetNamespace(), v.GetName(), since.String())
	}
	t.Print()
	return nil
}

func boxDelete(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	return sc.BoxDelete(context.Background(), namespace, name)
}

func boxTrigger(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	number, err := sc.BuildCreate(context.Background(), namespace, name)
	if err != nil {
		return err
	}
	writeString(strconv.FormatUint(number, 10))
	return nil
}

func buildGet(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("missing number")
	}
	number, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil || number < 1 {
		return errors.New("invalid number")
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, err := sc.BuildInfo(context.Background(), namespace, name, number)
	if err != nil {
		return err
	}

	b, err := yaml.Marshal(result)
	if err != nil {
		return err
	}
	write(b)
	return nil
}

func buildList(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, _, err := sc.BuildList(context.Background(), namespace, name, *getPage(cmd))
	if err != nil {
		return err
	}

	if len(result) == 0 {
		writeString("No resources found.")
		return nil
	}

	t := printer.NewTab("NUMBER", "PHASE", "STAGES", "STARTED", "STOPPED")
	for _, v := range result {
		var started, stopped string
		if v.Started > 0 {
			started = time.Unix(v.Started, 0).Format(time.DateTime)
		}
		if v.Stopped > 0 {
			stopped = time.Unix(v.Stopped, 0).Format(time.DateTime)
		}

		t.Add(
			strconv.FormatUint(v.Number, 10),
			v.Phase.String(),
			strconv.Itoa(len(v.Stages)),
			started,
			stopped,
		)
	}
	t.Print()
	return nil
}

func buildCancel(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("missing number")
	}
	number, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil || number < 1 {
		return errors.New("invalid number")
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	return sc.BuildCancel(context.Background(), namespace, name, number)
}
