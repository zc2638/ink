// Copyright © 2024 zc2638 <zc2638@qq.com>.
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

package runtime

import (
	"io"
	"strings"
)

// maskReplacer is an io.Writer that finds and masks sensitive data.
type maskReplacer struct {
	w io.WriteCloser
	r *strings.Replacer
}

// NewMaskReplacer returns a replacer that wraps io.Writer w.
func NewMaskReplacer(w io.WriteCloser, values []string) io.WriteCloser {
	var oldnew []string
	for _, v := range values {
		if len(v) == 0 {
			continue
		}

		for _, part := range strings.Split(v, "\n") {
			part = strings.TrimSpace(part)

			// avoid masking empty or single character strings.
			if len(part) < 2 {
				continue
			}

			masked := "******"
			oldnew = append(oldnew, part)
			oldnew = append(oldnew, masked)
		}
	}
	if len(oldnew) == 0 {
		return w
	}
	return &maskReplacer{
		w: w,
		r: strings.NewReplacer(oldnew...),
	}
}

// Write writes p to the base writer. The method scans for any
// sensitive data in p and masks before writing.
func (r *maskReplacer) Write(p []byte) (n int, err error) {
	_, err = r.w.Write([]byte(r.r.Replace(string(p))))
	return len(p), err
}

// Close closes the base writer.
func (r *maskReplacer) Close() error {
	return r.w.Close()
}
