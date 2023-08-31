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

package scheduler

import (
	"context"
	"testing"
	"time"
)

var noContext = context.Background()

func TestCollect(t *testing.T) {
	c := newCanceller()
	if err := c.Cancel(noContext, 1); err != nil {
		t.Errorf("Cancel failed")
	}
	if err := c.Cancel(noContext, 2); err != nil {
		t.Errorf("Cancel failed")
	}
	if err := c.Cancel(noContext, 3); err != nil {
		t.Errorf("Cancel failed")
	}
	if err := c.Cancel(noContext, 4); err != nil {
		t.Errorf("Cancel failed")
	}
	if err := c.Cancel(noContext, 5); err != nil {
		t.Errorf("Cancel failed")
	}
	c.canceled[3] = c.canceled[3].Add(time.Minute * -1)
	c.canceled[4] = time.Now().Add(time.Second * -1)
	c.canceled[5] = time.Now().Add(time.Second * -1)
	c.collect()

	if got, want := len(c.canceled), 3; got != want {
		t.Errorf("Want 3 canceled builds in the cache, got %d", got)
	}
	if _, ok := c.canceled[4]; ok {
		t.Errorf("Expect build id [4] removed")
	}
	if _, ok := c.canceled[5]; ok {
		t.Errorf("Expect build id [5] removed")
	}
}
