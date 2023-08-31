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

package v1

import "errors"

var ErrCanceled = errors.New("canceled")

type Phase string

func (s Phase) String() string {
	return string(s)
}

func ToPhase(s string) Phase {
	phase := Phase(s)
	switch phase {
	case PhaseWaiting,
		PhasePending,
		PhaseRunning,
		PhaseSucceeded,
		PhaseFailed,
		PhaseCanceled,
		PhaseSkipped:
	default:
		phase = PhaseUnknown
	}
	return phase
}

const (
	PhaseUnknown Phase = "Unknown"
	// PhaseWaiting used for dependencies,
	// the status description that waits for the dependency to finish executing.
	PhaseWaiting   Phase = "Waiting"
	PhasePending   Phase = "Pending"
	PhaseRunning   Phase = "Running"
	PhaseSucceeded Phase = "Succeeded"
	PhaseFailed    Phase = "Failed"
	PhaseCanceled  Phase = "Canceled"
	PhaseSkipped   Phase = "Skipped"
)

func (s Phase) IsDone() bool {
	switch s {
	case PhaseUnknown, PhaseWaiting, PhasePending, PhaseRunning:
		return false
	default:
		return true
	}
}

func (s Phase) IsSucceeded() bool {
	return s == PhaseSucceeded
}

func (s Phase) IsFailed() bool {
	return s == PhaseFailed
}
