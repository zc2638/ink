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

package clients

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"

	"github.com/zc2638/ink/core/constant"
)

func validateURI(v string) error {
	uri, err := url.Parse(v)
	if err != nil {
		return err
	}
	if uri.Scheme != "http" && uri.Scheme != "https" {
		return errors.New("invalid scheme")
	}
	if len(uri.Host) == 0 {
		return errors.New("invalid host")
	}
	return nil
}

func handleClientError(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}
	if resp.StatusCode() >= http.StatusBadRequest {
		b, err := io.ReadAll(resp.RawBody())
		if err != nil {
			return fmt.Errorf("read body error failed: %v", err)
		}
		body := strings.TrimSpace(string(b))
		errStr, err := strconv.Unquote(body)
		if err != nil {
			errStr = body
		}

		switch errStr {
		case constant.ErrNoRecord.Error():
			return constant.ErrNoRecord
		case constant.ErrAlreadyExists.Error():
			return constant.ErrAlreadyExists
		case context.Canceled.Error():
			return context.Canceled
		case context.DeadlineExceeded.Error():
			return context.DeadlineExceeded
		}
		return constant.NewHTTPError(resp.StatusCode(), errStr)
	}
	return nil
}
