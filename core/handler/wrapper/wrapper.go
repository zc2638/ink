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

package wrapper

import (
	"errors"
	"net/http"

	"github.com/99nil/gopkg/ctr"
	"github.com/go-chi/chi"
	"gorm.io/gorm"

	"github.com/zc2638/ink/core/constant"
)

func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func ErrorCode(w http.ResponseWriter, status int, v ...any) {
	var err error
	if len(v) > 0 {
		switch vv := v[0].(type) {
		case error:
			err = vv
		case string:
			err = errors.New(vv)
		}
	}
	if err == nil {
		w.WriteHeader(status)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = constant.ErrNoRecord
	}

	ctr.Logger().Errorln(err)
	ctr.JSON(w, err.Error(), status)
}

func BadRequest(w http.ResponseWriter, v ...any) {
	ErrorCode(w, http.StatusBadRequest, v...)
}

func InternalError(w http.ResponseWriter, v ...any) {
	ErrorCode(w, http.StatusInternalServerError, v...)
}
