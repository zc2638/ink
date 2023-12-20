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
	"strconv"
	"strings"
	"time"

	"github.com/99nil/gopkg/printer"
	"github.com/spf13/cobra"
	"github.com/zc2638/wslog"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/zc2638/ink/core/clients"
	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/worker"
	"github.com/zc2638/ink/core/worker/hooks"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/flags"
)

func NewCtl() *cobra.Command {
	cmd := &cobra.Command{
		Use: constant.CtlName,
	}
	persistentFlags := cmd.PersistentFlags()
	persistentFlags.AddGoFlag(
		flags.NewStringEnvFlag(constant.Name, "server", "http://localhost:2678",
			"the address and port of the inkd API server"),
	)
	persistentFlags.AddGoFlag(
		flags.NewBoolEnvFlag(constant.Name, "direct", false,
			"If true, the request will be executed directly by the built-in worker without request inkd"),
	)

	Register(cmd, "apply", "apply a configuration to a resource by file name", apply,
		flags.NewStringEnvFlag(constant.Name, "file", "",
			"that contains the configuration to apply"),
	)
	Register(cmd, "exec", "execute a configuration to a resource by file name", exec,
		flags.NewStringEnvFlag(constant.Name, "file", "",
			"that contains the configuration to exec"),
	)

	secretCmd := &cobra.Command{Use: "secret", Short: "secret operation"}
	Register(secretCmd, "list", "list secrets", secretList, secretListExample)
	Register(secretCmd, "delete", "delete secret", secretDelete, secretDeleteExample)

	workflowCmd := &cobra.Command{Use: "workflow", Short: "workflow operation"}
	Register(workflowCmd, "get", "get workflow info", workflowGet, workflowGetExample)
	Register(workflowCmd, "list", "list workflows", workflowList, workflowListExample)
	Register(workflowCmd, "delete", "delete workflow", workflowDelete, workflowDeleteExample)

	boxCmd := &cobra.Command{Use: "box", Short: "box operation"}
	Register(boxCmd, "get", "get box info", boxGet, boxGetExample)
	Register(boxCmd, "list", "list boxes", boxList, boxListExample)
	Register(boxCmd, "delete", "delete box", boxDelete, boxDeleteExample)
	boxTriggerCmd := Register(boxCmd, "trigger", "create a build for box", buildCreate, boxTriggerExample)
	boxTriggerCmd.Flags().StringArrayP("set", "s", nil, "setting values to workflow")

	buildCmd := &cobra.Command{Use: "build", Short: "build operation"}
	Register(buildCmd, "get", "get build info", buildGet, buildGetExample)
	Register(buildCmd, "list", "list builds", buildList, buildListExample)
	Register(buildCmd, "cancel", "cancel a build", buildCancel, buildCancelExample)
	buildCreateCmd := Register(buildCmd, "create", "create a build", buildCreate, buildCreateExample)
	buildCreateCmd.Flags().StringArrayP("set", "s", nil, "setting values to workflow")

	cmd.AddCommand(workflowCmd, boxCmd, buildCmd)
	return cmd
}

func secretList(cmd *cobra.Command, _ []string) error {
	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	result, err := sc.SecretList(context.Background(), v1.AllNamespace)
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

func secretDelete(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	return sc.SecretDelete(context.Background(), namespace, name)
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
	opt := v1.ListOption{
		Pagination: *getPage(cmd),
	}
	result, _, err := sc.WorkflowList(context.Background(), v1.AllNamespace, opt)
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
	opt := v1.ListOption{
		Pagination: *getPage(cmd),
	}
	result, _, err := sc.BoxList(context.Background(), v1.AllNamespace, opt)
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

func buildCreate(cmd *cobra.Command, args []string) error {
	namespace, name, err := getNN(args)
	if err != nil {
		return err
	}

	values, err := cmd.Flags().GetStringArray("set")
	if err != nil {
		return err
	}
	settings := make(map[string]string)
	for _, v := range values {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) < 2 {
			continue
		}
		settings[parts[0]] = parts[1]
	}

	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}
	number, err := sc.BuildCreate(context.Background(), namespace, name, settings)
	if err != nil {
		return err
	}
	writeString(strconv.FormatUint(number, 10))
	return nil
}

func apply(cmd *cobra.Command, _ []string) error {
	objSet, err := parseObjects(cmd)
	if err != nil {
		return err
	}
	sc, err := newServerClient(cmd)
	if err != nil {
		return err
	}

	ctx := context.Background()

	for _, obj := range objSet[v1.KindSecret] {
		var data v1.Secret
		if err := obj.ToObject(&data); err != nil {
			return err
		}
		_, err := sc.SecretInfo(ctx, data.GetNamespace(), data.GetName())
		if err == nil {
			if err = sc.SecretUpdate(ctx, &data); err == nil {
				writeString(fmt.Sprintf("Update: %s", data.Metadata.String()))
			}
		} else if errors.Is(err, constant.ErrNoRecord) {
			if err = sc.SecretCreate(ctx, &data); err == nil {
				writeString(fmt.Sprintf("Create: %s", data.Metadata.String()))
			}
		}
		if err != nil {
			return err
		}
	}
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

func exec(cmd *cobra.Command, _ []string) error {
	objSet, err := parseObjects(cmd)
	if err != nil {
		return err
	}
	dataCh := make(chan *v1.Data)
	wc := clients.NewClientDirect(dataCh)

	allSecrets := make([]*v1.Secret, 0)
	for _, obj := range objSet[v1.KindSecret] {
		var secret v1.Secret
		if err := obj.ToObject(&secret); err != nil {
			return err
		}
		allSecrets = append(allSecrets, &secret)
	}

	// TODO 暂时仅支持 workflow 单独串行执行
	for _, obj := range objSet[v1.KindWorkflow] {
		var workflow v1.Workflow
		if err := obj.ToObject(&workflow); err != nil {
			return err
		}

		var hook worker.Hook
		workerObj := workflow.Worker()
		switch workerObj.Kind {
		case v1.WorkerKindHost:
			hook, err = hooks.NewHost()
			if err != nil {
				return fmt.Errorf("init host hook failed: %v", err)
			}
		case v1.WorkerKindDocker:
			hook, err = hooks.NewDocker("", "")
			if err != nil {
				return fmt.Errorf("init docker hook failed: %v", err)
			}
		default:
			return fmt.Errorf("unsupported kind: %s", workerObj.Kind)
		}

		secrets := make([]*v1.Secret, 0)
		for _, sec := range allSecrets {
			if sec.GetNamespace() != workflow.GetNamespace() {
				continue
			}
			secrets = append(secrets, sec)
		}

		data := &v1.Data{
			Workflow: &workflow,
			Status: &v1.Stage{
				Number:    1,
				Phase:     v1.PhasePending,
				Name:      workflow.GetName(),
				Limit:     workflow.Spec.Concurrency,
				Worker:    *workflow.Worker(),
				DependsOn: workflow.Spec.DependsOn,
			},
			Secrets: secrets,
		}
		for sk, sv := range workflow.Spec.Steps {
			step := &v1.Step{
				Number: uint64(sk) + 1,
				Phase:  v1.PhasePending,
				Name:   sv.Name,
			}
			step.ID = step.Number
			data.Status.Steps = append(data.Status.Steps, step)
		}

		log := wslog.With(
			"name", workflow.GetName(),
			"namespace", workflow.GetNamespace(),
			"worker_kind", workerObj.Kind.String(),
		)
		log.Info("Begin workflow")

		eg, ctx := errgroup.WithContext(context.Background())
		eg.Go(func() error { return worker.Run(ctx, wc, hook) })
		eg.Go(func() error {
			select {
			case dataCh <- data:
			case <-ctx.Done():
			}
			return nil
		})
		if err := eg.Wait(); err != nil {
			return err
		}
		log.Info("End workflow")
	}
	return nil
}
