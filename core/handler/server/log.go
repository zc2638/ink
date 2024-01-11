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
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/99nil/gopkg/ctr"
	"github.com/99nil/gopkg/sse"

	"github.com/zc2638/ink/core/handler/wrapper"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
	"github.com/zc2638/ink/pkg/livelog"
)

func logInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")
		number, _ := strconv.ParseUint(
			wrapper.URLParam(r, "number"), 10, 64)
		stageNumber, _ := strconv.ParseUint(
			wrapper.URLParam(r, "stage"), 10, 64)
		stepNumber, _ := strconv.ParseUint(
			wrapper.URLParam(r, "step"), 10, 64)
		db := database.FromRequest(r)

		boxS := &storageV1.Box{Namespace: namespace, Name: name}
		if err := db.Where(boxS).First(boxS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		buildS := &storageV1.Build{BoxID: boxS.ID, Number: number}
		if err := db.Where(buildS).First(buildS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		stageS := &storageV1.Stage{BuildID: buildS.ID, Number: stageNumber}
		if err := db.Where(stageS).First(stageS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		stepS := &storageV1.Step{StageID: stageS.ID, Number: stepNumber}
		if err := db.Where(stepS).First(stepS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		logS := new(storageV1.Log)
		logS.SetID(stepS.ID)
		if err := db.Where(logS).First(logS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Bytes(w, logS.Data)
	}
}

func logWatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := wrapper.URLParam(r, "namespace")
		name := wrapper.URLParam(r, "name")
		number, _ := strconv.ParseUint(
			wrapper.URLParam(r, "number"), 10, 64)
		stageNumber, _ := strconv.ParseUint(
			wrapper.URLParam(r, "stage"), 10, 64)
		stepNumber, _ := strconv.ParseUint(
			wrapper.URLParam(r, "step"), 10, 64)
		db := database.FromRequest(r)

		boxS := &storageV1.Box{Namespace: namespace, Name: name}
		if err := db.Where(boxS).First(boxS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		buildS := &storageV1.Build{BoxID: boxS.ID, Number: number}
		if err := db.Where(buildS).First(buildS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		stageS := &storageV1.Stage{BuildID: buildS.ID, Number: stageNumber}
		if err := db.Where(stageS).First(stageS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		stepS := &storageV1.Step{StageID: stageS.ID, Number: stepNumber}
		if err := db.Where(stepS).First(stepS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}

		sender, err := sse.NewSender(w)
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		sender.Ping()

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		ll := livelog.FromRequest(r)
		lineCh, closeCh, err := ll.Watch(ctx, strconv.FormatUint(stepS.ID, 10))
		if err != nil {
			sender.SendError("", err)
			return
		}
		if closeCh == nil {
			sender.Close()
			return
		}

		pingChan := time.After(30 * time.Second)
		timeoutChan := time.After(24 * time.Hour)
	L:
		for {
			select {
			case <-ctx.Done():
				break L
			case <-closeCh:
				break L
			case <-timeoutChan:
				break L
			case <-pingChan:
				sender.Ping()
			case line := <-lineCh:
				data, err := json.Marshal(line)
				if err != nil {
					sender.SendError("", err)
					continue
				}
				sender.SendMessage(&sse.Message{
					Event: "data",
					Data:  string(data),
				})
			}
		}
		sender.Close()
	}
}
