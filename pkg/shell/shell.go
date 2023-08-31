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

package shell

import (
	"bytes"
	"fmt"
	"strings"
)

// Suffix provides the shell script suffix. For posix systems
// this value is an empty string.
const Suffix = ""

const scriptFormat = `
echo + %s
%s
`

// Command returns the shell command and arguments.
func Command() (string, []string) {
	return "/bin/sh", []string{"-c"}
}

func EchoEnvCommand(env string, command string) []string {
	if command == "" {
		command = "/bin/sh"
	}
	return []string{fmt.Sprintf(`echo "$%s" | %s`, env, command)}
}

// Script converts a slice of individual shell commands to
// a posix-compliant shell script.
func Script(commands []string) string {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf)
	fmt.Fprintf(buf, "set -e")
	fmt.Fprintln(buf)
	for _, command := range commands {
		escaped := fmt.Sprintf("%q", command)
		escaped = strings.ReplaceAll(escaped, "$", `\$`)
		buf.WriteString(fmt.Sprintf(scriptFormat, escaped, command))
	}
	return buf.String()
}
