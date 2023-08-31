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

package describe

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/zc2638/ink/core/constant"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/livelog"
	"github.com/zc2638/ink/test/e2e/framework"
)

func createOrUpdateStage(ctx context.Context, stage *v1.Stage) {
	_, err := framework.Server.StageInfo(ctx, stage.Namespace, stage.Name)
	if errors.Is(err, constant.ErrNoRecord) {
		err = framework.Server.StageCreate(ctx, stage)
		gomega.Expect(err).To(gomega.BeNil(), "create stage failed")
		return
	}
	gomega.Expect(err).To(gomega.BeNil(), "create stage failed")
	err = framework.Server.StageUpdate(ctx, stage)
	gomega.Expect(err).To(gomega.BeNil(), "update stage failed")
}

func createOrUpdateBox(ctx context.Context, box *v1.Box) {
	_, err := framework.Server.BoxInfo(ctx, box.Namespace, box.Name)
	if errors.Is(err, constant.ErrNoRecord) {
		err = framework.Server.BoxCreate(ctx, box)
		gomega.Expect(err).To(gomega.BeNil(), "create box failed")
		return
	}
	gomega.Expect(err).To(gomega.BeNil(), "create box failed")
	err = framework.Server.BoxUpdate(ctx, box)
	gomega.Expect(err).To(gomega.BeNil(), "update box failed")
}

var _ = ginkgo.Describe("Complete process", func() {
	ginkgo.It("success", func() {
		ctx := context.Background()

		ginkgo.By("[Server] Create or update Stage")
		stage := &v1.Stage{
			Metadata: v1.Metadata{
				Kind:      v1.KindStage,
				Name:      framework.BuildName("test"),
				Namespace: v1.DefaultNamespace,
			},
			Spec: v1.StageSpec{
				Worker: &v1.Worker{Kind: v1.WorkerKindDocker},
				Steps: []v1.Step{
					{
						Name:    "step1",
						Image:   "alpine:3.18",
						Command: []string{"echo", "step1"},
					},
					{
						Name:    "step1",
						Image:   "alpine:3.18",
						Command: []string{"echo", "step2"},
					},
				},
			},
		}
		createOrUpdateStage(ctx, stage)

		ginkgo.By("[Server] Create or update Box")
		box := &v1.Box{
			Metadata: v1.Metadata{
				Kind:      v1.KindBox,
				Name:      framework.BuildName("test"),
				Namespace: v1.DefaultNamespace,
			},
			Resources: []v1.BoxResource{
				{Kind: v1.KindStage, Name: stage.Name},
			},
		}
		createOrUpdateBox(ctx, box)

		ginkgo.By("[Server] Create build")
		_, err := framework.Server.BuildCreate(ctx, box.Namespace, box.Name)
		gomega.Expect(err).To(gomega.BeNil(), "create build failed")
	})
})

func watchLog(logCh <-chan *livelog.Line, logErrCh <-chan error, expectCount int) error {
	var count int
	for {
		select {
		case line := <-logCh:
			fmt.Printf("%+v\n", line)
			count++
			if expectCount == count {
				return nil
			}
		case err := <-logErrCh:
			if err == nil || errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
