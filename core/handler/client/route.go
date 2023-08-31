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

package client

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Handler(middlewares chi.Middlewares) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Recoverer,
		middleware.NoCache,
		middleware.Logger,
	)
	r.Use(middlewares...)

	r.Post("/status", handleStatus)
	r.Post("/stage", handleRequest)
	r.Post("/stage/{stage}", handleAccept())
	r.Get("/stage/{stage}", handleInfo())
	r.Post("/stage/{stage}/begin", handleStageBegin())
	r.Post("/stage/{stage}/end", handleStageEnd())
	r.Post("/step/{step}/begin", handleStepBegin())
	r.Post("/step/{step}/end", handleStepEnd())
	r.Post("/step/{step}/logs/upload", handleLogUpload())
	r.Post("/build/{build}/watch", handleWatchCancel)
	return r
}
