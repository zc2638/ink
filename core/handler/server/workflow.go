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

func workflowList(workflowSrv service.Workflow) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		page := v1.GetPagination(r)
		result, err := workflowSrv.List(r.Context(), namespace, page)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, page.List(result))
	}
}

func workflowInfo(workflowSrv service.Workflow) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		result, err := workflowSrv.Info(r.Context(), namespace, name)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, result)
	}
}

func workflowCreate(workflowSrv service.Workflow) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in v1.Workflow
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			wrapper.BadRequest(w, err)
			return
		}

		if err := workflowSrv.Create(r.Context(), &in); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

func workflowUpdate(workflowSrv service.Workflow) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		var in v1.Workflow
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		in.SetNamespace(namespace)
		in.SetName(name)

		if err := workflowSrv.Update(r.Context(), &in); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

func workflowDelete(workflowSrv service.Workflow) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		if err := workflowSrv.Delete(r.Context(), namespace, name); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}
