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

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

// Interface schedules Build stages for execution.
type Interface interface {
	// Schedule schedules the stage for execution.
	Schedule(context.Context)

	// Request requests the next stage scheduled for execution.
	Request(context.Context, v1.Worker) (*v1.StageStatus, error)

	// Cancel cancels scheduled or running jobs associated
	// with the parent build ID.
	Cancel(context.Context, int64) error

	// Canceled blocks and listens for a cancellation event and
	// returns true if the build has been canceled.
	Canceled(context.Context, int64) (bool, error)

	// Pause pauses the scheduler and prevents new stages
	// from being scheduled for execution.
	Pause(context.Context) error

	// Resume unpauses the scheduler, allowing new stages
	// to be scheduled for execution.
	Resume(context.Context) error
}

// New creates a new scheduler.
func New(storeFunc StoreFunc) Interface {
	return scheduler{
		queue:     newQueue(storeFunc),
		canceller: newCanceller(),
	}
}

type scheduler struct {
	*queue
	*canceller
}
