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

type Example string

func (s Example) String() string {
	return string(s)
}

const secretListExample Example = `
# List secrets
inkctl secret list
`

const secretDeleteExample Example = `
# Definition
inkctl secret delete {namespace}/{name}

# Delete a secret
inkctl secret delete default/test

# Delete secret with default namespace
inkctl secret delete test
`

const workflowListExample Example = `
# List workflows (default size: 10)
inkctl workflow list

# List all workflows
inkctl workflow list --size -1

# List workflows specify the page and size
inkctl workflow list --size 15 --page 2
`

const workflowGetExample Example = `
# Definition
inkctl workflow get {namespace}/{name}

# Get workflow info
inkctl workflow get default/test

# Get workflow info with default namespace
inkctl workflow get test
`

const workflowDeleteExample Example = `
# Definition
inkctl workflow delete {namespace}/{name}

# Delete a workflow
inkctl workflow delete default/test

# Delete workflow with default namespace
inkctl workflow delete test
`

const boxListExample Example = `
# List boxes (default size: 10)
inkctl box list

# List all boxes
inkctl box list --size -1

# List boxes specify the page and size
inkctl box list --size 15 --page 2
`

const boxGetExample Example = `
# Definition
inkctl box get {namespace}/{name}

# Get box info
inkctl box get default/test

# Get box info with default namespace
inkctl box get test
`

const boxDeleteExample Example = `
# Definition
inkctl box delete {namespace}/{name}

# Delete box info
inkctl box delete default/test

# Delete box info with default namespace
inkctl box delete test
`

const boxTriggerExample Example = `
# Definition
inkctl box trigger {namespace}/{name}

# Trigger box to create a build
inkctl box trigger default/test

# Trigger box to create a build with default namespace
inkctl box trigger test
`

const buildListExample Example = `
# Definition
inkctl build list {namespace}/{name}

# List builds (default size: 10)
inkctl build list default/test

# List all builds
inkctl build list --size -1

# List builds specify the page and size
inkctl build list --size 15 --page 2
`

const buildGetExample Example = `
# Definition
inkctl build get {namespace}/{name} {number}

# Get build info
inkctl build get default/test 1

# Get build info with default namespace
inkctl build get test 1
`

const buildCancelExample Example = `
# Definition
inkctl build cancel {namespace}/{name} {number}

# Cancel a build
inkctl build cancel default/test 1

# Cancel a build with default namespace
inkctl build cancel test 1
`

const buildCreateExample Example = `
# Definition
inkctl build create {namespace}/{name}

# Create a build for the box
inkctl build create default/test

# Create a build for the box with default namespace
inkctl build create test
`
