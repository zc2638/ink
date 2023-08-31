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

package livelog

import (
	"context"
	"net/http"
)

type key struct{}

// WithContext returns a new context with the provided livelog.
func WithContext(ctx context.Context, ins Interface) context.Context {
	return context.WithValue(ctx, key{}, ins)
}

// FromContext retrieves the current livelog from the context. If no
// livelog is available, the nil value is returned.
func FromContext(ctx context.Context) Interface {
	v := ctx.Value(key{})
	if v == nil {
		return nil
	}
	return v.(Interface)
}

// FromRequest retrieves the current livelog from the request. If no
// livelog is available, the nil value is returned.
func FromRequest(r *http.Request) Interface {
	return FromContext(r.Context())
}
