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

package selector

import (
	"fmt"
	"slices"
	"strconv"
)

type Selector struct {
	Matches    Match       `json:"matches,omitempty" yaml:"matches,omitempty"`
	Operations []Operation `json:"operations,omitempty" yaml:"operations,omitempty"`
}

func (s *Selector) Validate() error {
	for _, operation := range s.Operations {
		if err := operation.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector) Match(data map[string]string) bool {
	if s == nil {
		return true
	}
	if s.Matches != nil && !s.Matches.Match(data) {
		return false
	}
	for _, operation := range s.Operations {
		if !operation.Match(data) {
			return false
		}
	}
	return true
}

type Match map[string]string

func (m Match) Match(data map[string]string) bool {
	for k, v := range m {
		dv, ok := data[k]
		if !ok || dv != v {
			return false
		}
	}
	return true
}

type Operation struct {
	Key      string   `json:"key" yaml:"key"`
	Operator Operator `json:"operator" yaml:"operator"`
	Values   []string `json:"values" yaml:"values"`
}

// Validate Operation.
// If any of these rules is violated, an error is returned:
//  1. The operator can only be In, NotIn, Equals, DoubleEquals, Gt, Lt, NotEquals, Exists, or DoesNotExist.
//  2. If the operator is In or NotIn, the values set must be non-empty.
//  3. If the operator is Equals, DoubleEquals, or NotEquals, the values set must contain one value.
//  4. If the operator is Exists or DoesNotExist, the value set must be empty.
//  5. If the operator is Gt or Lt, the values set must contain only one value, which will be interpreted as an integer.
//  6. The key is invalid due to its length, or sequence of characters. See validateLabelKey for more details.
//
// The empty string is a valid value in the input values set.
func (o *Operation) Validate() error {
	key := o.Key
	op := o.Operator
	vals := o.Values
	switch op {
	case In, NotIn:
		if len(vals) == 0 {
			return fmt.Errorf("%s: for 'in', 'notin' operators, values set can't be empty", key)
		}
	case Equals, NotEquals:
		if len(vals) != 1 {
			return fmt.Errorf("%s: exact-match compatibility requires one single value", key)
		}
	case Exists, DoesNotExist:
		if len(vals) != 0 {
			return fmt.Errorf("%s: values set must be empty for exists and does not exist", key)
		}
	case GreaterThan, LessThan:
		if len(vals) != 1 {
			return fmt.Errorf("%s: for 'Gt', 'Lt' operators, exactly one value is required", key)
		}
		for i := range vals {
			if _, err := strconv.ParseInt(vals[i], 10, 64); err != nil {
				return fmt.Errorf("%s: for 'Gt', 'Lt' operators, the value must be an integer", key)
			}
		}
	default:
		return fmt.Errorf("operator not support: %v", op)
	}
	return nil
}

func (o *Operation) Match(data map[string]string) bool {
	val, ok := data[o.Key]
	if !ok {
		return o.Operator == DoesNotExist
	}
	switch o.Operator {
	case In:
		return slices.Contains(o.Values, val)
	case NotIn:
		return !slices.Contains(o.Values, val)
	case Equals:
		return val == o.Values[0]
	case NotEquals:
		return val != o.Values[0]
	case Exists:
		return true
	case GreaterThan:
		vn1, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return false
		}
		vn2, err := strconv.ParseInt(o.Values[0], 10, 64)
		if err != nil {
			return false
		}
		return vn1 > vn2
	case LessThan:
		vn1, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return false
		}
		vn2, err := strconv.ParseInt(o.Values[0], 10, 64)
		if err != nil {
			return false
		}
		return vn1 < vn2
	}
	return false
}
