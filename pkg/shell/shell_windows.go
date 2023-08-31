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

// Suffix provides the shell script suffix.
const Suffix = ".ps1"

// optionScript is a helper script this is added to the build
// to set shell options, in this case, to exit on error.
const optionScript = `$erroractionpreference = "stop"`

const scriptFormat = `
echo %s
%s
if ($LastExitCode -gt 0) { exit $LastExitCode }
`

// Command returns the Powershell command and arguments.
func Command() (string, []string) {
	return "powershell", []string{
		"-noprofile",
		"-noninteractive",
		"-command",
	}
}

func EchoEnvCommand(env string, command string) []string {
	if command == "" {
		command = "iex"
	}
	return []string{fmt.Sprintf("echo $Env:%s | %s", env, command)}
}

// Script converts a slice of individual shell commands to
// a powershell script.
func Script(commands []string) string {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf)
	fmt.Fprintf(buf, optionScript)
	fmt.Fprintln(buf)
	for _, command := range commands {
		escaped := fmt.Sprintf("%q", "+ "+command)
		escaped = strings.Replace(escaped, "$", "`$", -1)
		buf.WriteString(fmt.Sprintf(traceScript, escaped, command))
	}
	return buf.String()
}
