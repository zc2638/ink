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

func secretList(secretSrv service.Secret) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		page := v1.GetPagination(r)
		result, err := secretSrv.List(r.Context(), namespace, v1.ListOption{
			Pagination:    *page,
			LabelSelector: r.URL.Query().Get("labelSelector"),
		})
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, result)
	}
}

func secretInfo(secretSrv service.Secret) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		result, err := secretSrv.Info(r.Context(), namespace, name)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, result)
	}
}

func secretCreate(secretSrv service.Secret) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in v1.Secret
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			wrapper.BadRequest(w, err)
			return
		}

		if err := secretSrv.Create(r.Context(), &in); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

func secretUpdate(secretSrv service.Secret) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		var in v1.Secret
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		in.SetNamespace(namespace)
		in.SetName(name)

		if err := secretSrv.Update(r.Context(), &in); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

func secretDelete(secretSrv service.Secret) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")

		if err := secretSrv.Delete(r.Context(), namespace, name); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}
