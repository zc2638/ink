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

package command

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func ParseAllConfig(configPath string, data any, envPrefix string, subKey string) (*viper.Viper, error) {
	if len(configPath) == 0 {
		return nil, errors.New("config path must be defined")
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	if len(subKey) > 0 {
		v = v.Sub(subKey)
		if v == nil {
			return nil, fmt.Errorf("sub key '%s' not found in config file", subKey)
		}
	}

	return v, v.Unmarshal(data, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
}

func ParseConfig(configPath string, data any, envPrefix string) (*viper.Viper, error) {
	return ParseAllConfig(configPath, data, envPrefix, "")
}

func DefaultConfig(service string) string {
	var value string
	if len(service) > 0 {
		value = os.Getenv(strings.ToUpper(service + "_config"))
	}
	if len(value) == 0 {
		value = fmt.Sprintf("config/%s.yaml", service)
	}
	return value
}
