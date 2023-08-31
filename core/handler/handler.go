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

package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/99nil/gopkg/ctr"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/segmentio/ksuid"
	"github.com/zc2638/wslog"
	"gorm.io/gorm"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/handler/client"
	"github.com/zc2638/ink/core/handler/server"
	"github.com/zc2638/ink/core/scheduler"
	"github.com/zc2638/ink/pkg/database"
	"github.com/zc2638/ink/pkg/livelog"
)

var corsOpts = cors.Options{
	AllowedOrigins: []string{"*"},
	AllowedMethods: []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
	},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	ExposedHeaders:   []string{"Link"},
	AllowCredentials: true,
	MaxAge:           300,
}

func New(log *wslog.Logger, db *gorm.DB, ll livelog.Interface, sched scheduler.Interface) http.Handler {
	apiMiddlewares := chi.Middlewares{
		cors.New(corsOpts).Handler,
		serviceMiddleware(log, ll, sched, db),
		timeoutMiddleware,
	}

	mux := chi.NewMux()
	mux.Get("/", func(w http.ResponseWriter, r *http.Request) { ctr.OK(w, "Hello Ink") })
	mux.Mount("/api/core/v1", server.Handler(apiMiddlewares))
	mux.Mount("/api/client/v1", client.Handler(apiMiddlewares))
	return mux
}

func serviceMiddleware(log *wslog.Logger, ll livelog.Interface, sched scheduler.Interface, db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = ksuid.New().String()
			}
			log = log.With("request-id", id)

			ctx := r.Context()
			ctx = wslog.WithContext(ctx, log)
			ctx = livelog.WithContext(ctx, ll)
			ctx = scheduler.WithContext(ctx, sched)
			ctx = database.WithContext(ctx, db)

			if !log.Enabled(slog.LevelDebug) {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r.WithContext(ctx))
			end := time.Now()
			log.With(
				"method", r.Method,
				"status", ww.Status(),
				"request", r.RequestURI,
				"remote", r.RemoteAddr,
				"latency", end.Sub(start),
				"time", end.Format(time.RFC3339),
			)
		})
	}
}

func timeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), constant.DefaultHTTPTimeout)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
