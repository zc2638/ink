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
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/docker/distribution/reference"
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

const DockerConfigJSONKey = ".dockerconfigjson"

type DockerAuths struct {
	Auths map[string]DockerAuth `json:"auths" yaml:"auths"`
}

func (das *DockerAuths) Match(image string) string {
	for k, v := range das.Auths {
		if !das.matchHostname(image, k) {
			continue
		}

		authMap := map[string]string{
			"username": v.Username,
			"password": v.Password,
		}
		bs, err := json.Marshal(&authMap)
		if err != nil {
			return ""
		}
		return base64.URLEncoding.EncodeToString(bs)
	}
	return ""
}

func (das *DockerAuths) matchHostname(image, hostname string) bool {
	ref, err := reference.ParseAnyReference(image)
	if err != nil {
		return false
	}
	named, err := reference.ParseNamed(ref.String())
	if err != nil {
		return false
	}
	if hostname == "index.docker.io" {
		hostname = "docker.io"
	}
	// the auth address could be a fully qualified url in which case,
	// we should parse so we can extract the domain name.
	if strings.HasPrefix(hostname, "http://") ||
		strings.HasPrefix(hostname, "https://") {
		parsed, err := url.Parse(hostname)
		if err == nil {
			hostname = parsed.Host
		}
	}
	return reference.Domain(named) == hostname
}

type DockerAuth struct {
	Auth     string `json:"auth" yaml:"auth"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

func (da *DockerAuth) Encrypt() {
	authStr := da.Username + ":" + da.Password
	da.Auth = base64.StdEncoding.EncodeToString([]byte(authStr))
}

func (da *DockerAuth) Decrypt() error {
	b, err := base64.StdEncoding.DecodeString(da.Auth)
	if err != nil {
		return err
	}

	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) < 2 {
		return errors.New("invalid auth")
	}
	da.Username = parts[0]
	da.Password = parts[1]
	return nil
}
