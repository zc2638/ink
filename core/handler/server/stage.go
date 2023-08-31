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
	"net/http"

	"github.com/99nil/gopkg/ctr"

	"github.com/zc2638/ink/core/handler/wrapper"
	"github.com/zc2638/ink/core/service"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

func stageList(stageSrv service.Stage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := v1.GetPagination(r)
		result, err := stageSrv.List(r.Context(), page)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, page.List(result))
	}
}

func stageInfo(stageSrv service.Stage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		result, err := stageSrv.Info(r.Context(), namespace, name)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, result)
	}
}

func stageCreate(stageSrv service.Stage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in v1.Stage
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			wrapper.BadRequest(w, err)
			return
		}

		if err := stageSrv.Create(r.Context(), &in); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

func stageUpdate(stageSrv service.Stage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		var in v1.Stage
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		in.SetNamespace(namespace)
		in.SetName(name)

		if err := stageSrv.Update(r.Context(), &in); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

func stageDelete(stageSrv service.Stage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		if err := stageSrv.Delete(r.Context(), namespace, name); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}
