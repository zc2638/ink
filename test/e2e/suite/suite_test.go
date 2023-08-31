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

package suite_test

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/zc2638/wslog"

	"github.com/zc2638/ink/core/clients"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/test/e2e/framework"

	_ "github.com/zc2638/ink/test/e2e/describe"
)

func Test(t *testing.T) {
	addr := "http://localhost:2678"
	serverC, err := clients.NewServer(addr)
	if err != nil {
		t.Fatal(err)
	}
	clientC, err := clients.NewClient(addr, framework.Name, &v1.Worker{Kind: v1.WorkerKindHost})
	if err != nil {
		t.Fatal(err)
	}
	framework.Addr = addr
	framework.Server = serverC.V1()
	framework.Client = clientC.V1()

	gomega.RegisterFailHandler(ginkgo.Fail)
	suiteConfig, reporterConfig := framework.CreateGinkgoConfig()
	wslog.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunID, suiteConfig.ParallelProcess)
	ginkgo.RunSpecs(t, "Ink e2e suite", suiteConfig, reporterConfig)
}
