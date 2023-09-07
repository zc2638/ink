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
	"net/http"

	"github.com/go-chi/chi"

	"github.com/zc2638/ink/core/service/box"
	"github.com/zc2638/ink/core/service/build"
	"github.com/zc2638/ink/core/service/secret"
	"github.com/zc2638/ink/core/service/workflow"
)

func Handler(middlewares chi.Middlewares) http.Handler {
	r := chi.NewRouter()
	r.Use(middlewares...)

	workflowSrv := workflow.New()
	boxSrv := box.New()
	buildSrv := build.New()
	secretSrv := secret.New()

	r.Route("/box", func(r chi.Router) {
		r.Get("/", boxList(boxSrv))
		r.Post("/", boxCreate(boxSrv))

		r.Route("/{namespace}/{name}", func(r chi.Router) {
			r.Get("/", boxInfo(boxSrv))
			r.Put("/", boxUpdate(boxSrv))
			r.Delete("/", boxDelete(boxSrv))

			r.Route("/build", func(r chi.Router) {
				r.Get("/", buildList(buildSrv))
				r.Post("/", buildCreate(buildSrv))

				r.Route("/{number}", func(r chi.Router) {
					r.Get("/", buildInfo(buildSrv))
					r.Post("/cancel", buildCancel(buildSrv))
					r.Get("/logs/{stage}/{step}", logInfo())
					r.Post("/logs/{stage}/{step}", logWatch())
				})
			})
		})
	})

	r.Route("/workflow", func(r chi.Router) {
		r.Get("/", workflowList(workflowSrv))
		r.Post("/", workflowCreate(workflowSrv))
		r.Route("/{namespace}/{name}", func(r chi.Router) {
			r.Get("/", workflowInfo(workflowSrv))
			r.Put("/", workflowUpdate(workflowSrv))
			r.Delete("/", workflowDelete(workflowSrv))
		})
	})

	r.Route("/secret", func(r chi.Router) {
		r.Get("/", secretList(secretSrv))
		r.Post("/", secretCreate(secretSrv))
		r.Route("/{namespace}/{name}", func(r chi.Router) {
			r.Get("/", secretInfo(secretSrv))
			r.Put("/", secretUpdate(secretSrv))
			r.Delete("/", secretDelete(secretSrv))
		})
	})
	return r
}
