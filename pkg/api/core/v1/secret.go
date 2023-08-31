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

package v1

import (
	"encoding/base64"
	"fmt"
)

type Secret struct {
	Metadata `yaml:",inline"`

	Data        map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	EncryptData map[string]string `json:"encryptData,omitempty" yaml:"encryptData,omitempty"`
}

func (s *Secret) Encrypt() {
	if s.EncryptData == nil {
		s.EncryptData = make(map[string]string)
	}
	for k, v := range s.Data {
		val := base64.StdEncoding.EncodeToString([]byte(v))
		s.EncryptData[k] = val
	}
}

func (s *Secret) Decrypt() error {
	if s.Data == nil {
		s.Data = make(map[string]string)
	}
	for k, v := range s.EncryptData {
		valBytes, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return fmt.Errorf("decryption key(%s) value failed: %v", k, err)
		}
		s.Data[k] = string(valBytes)
	}
	return nil
}
