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

package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/99nil/gopkg/ctr"

	"github.com/zc2638/ink/core/handler/wrapper"
	"github.com/zc2638/ink/core/scheduler"
	"github.com/zc2638/ink/core/service"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

func buildList(buildSrv service.Build) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")
		page := v1.GetPagination(r)

		result, err := buildSrv.List(r.Context(), namespace, name, page)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, page.List(result))
	}
}

func buildInfo(buildSrv service.Build) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")
		number, _ := strconv.ParseUint(
			wrapper.URLParam(r, "number"), 10, 64)
		if number == 0 {
			wrapper.BadRequest(w, errors.New("invalid build number"))
			return
		}

		result, err := buildSrv.Info(r.Context(), namespace, name, number)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, result)
	}
}

func buildCreate(buildSrv service.Build) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		settings := make(map[string]string)
		_ = json.NewDecoder(r.Body).Decode(&settings)
		query := r.URL.Query()
		for k := range query {
			settings[k] = query.Get(k)
		}

		number, err := buildSrv.Create(r.Context(), namespace, name, settings)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}

		sched := scheduler.FromRequest(r)
		sched.Schedule(r.Context())
		ctr.OK(w, number)
	}
}

func buildCancel(buildSrv service.Build) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")
		number, _ := strconv.ParseUint(
			wrapper.URLParam(r, "number"), 10, 64)
		if number == 0 {
			wrapper.BadRequest(w, errors.New("invalid build number"))
			return
		}

		if err := buildSrv.Cancel(r.Context(), namespace, name, number); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}
