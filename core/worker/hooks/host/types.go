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
	"os"
	"path/filepath"

	"github.com/zc2638/ink/core/worker"
)

func getRootDir(stage *worker.Stage) string {
	return filepath.Join(os.TempDir(), stage.ID)
}

func getHomedir(stage *worker.Stage) string {
	dir := getRootDir(stage)
	return filepath.Join(dir, "/home/ink")
}
