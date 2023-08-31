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

package constant

import (
	"errors"
	"fmt"
	"time"
)

const (
	Name       = "ink"
	CtlName    = "inkctl"
	DaemonName = "inkd"
	WorkerName = "inker"
)

const WorkspacePath = "/ink/src"

// DefaultWaitTime defines the default waiting time(second)
const DefaultWaitTime = 60

// DefaultHTTPTimeout defines the default http request timeout
var DefaultHTTPTimeout = time.Second * 30

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNoRecord      = errors.New("no record")
	ErrInvalidName   = errors.New("invalid name")
)

func NewHTTPError(code int, msg string) *HTTPError {
	return &HTTPError{Code: code, Err: msg}
}

type HTTPError struct {
	Code int
	Err  string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("code: %d, error: %s", e.Code, e.Err)
}
