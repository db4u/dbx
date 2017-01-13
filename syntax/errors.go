// Copyright (C) 2016 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syntax

import (
	"fmt"
	"strings"
	"text/scanner"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
)

var errorPosition = errors.GenSym()

func setErrorPosition(pos scanner.Position) errors.ErrorOption {
	return errors.SetData(errorPosition, pos)
}

func getErrorPosition(err error) *scanner.Position {
	pos, ok := errors.GetData(err, errorPosition).(scanner.Position)
	if ok {
		return &pos
	}
	return nil
}

func expectedKeyword(pos scanner.Position, actual string, expected ...string) (
	err error) {

	if len(expected) == 1 {
		return Error.NewWith(fmt.Sprintf("%s: expected %q, got %q",
			pos, expected[0], actual),
			setErrorPosition(pos))
	} else {
		return Error.NewWith(fmt.Sprintf("%s: expected one of %q, got %q",
			pos, expected, actual),
			setErrorPosition(pos))
	}
}

func expectedToken(pos scanner.Position, actual Token, expected ...Token) (
	err error) {

	if len(expected) == 1 {
		return Error.NewWith(fmt.Sprintf("%s: expected %q; got %q",
			pos, expected[0], actual),
			setErrorPosition(pos))
	} else {
		return Error.NewWith(fmt.Sprintf("%s: expected one of %v; got %q",
			pos, expected, actual),
			setErrorPosition(pos))
	}
}

func errorAt(n node, format string, args ...interface{}) error {
	return Error.NewWith(
		fmt.Sprintf("%s: %s", n.getPos(), fmt.Sprintf(format, args...)),
		setErrorPosition(n.getPos()))
}

func previouslyDefined(n node, kind, field string,
	where scanner.Position) error {

	return errorAt(n, "%s already defined on %s. previous definition at %s",
		field, kind, where)
}

func flagField(kind, field string, val **ast.Bool) func(*tupleNode) error {
	return func(node *tupleNode) error {
		if *val != nil {
			return previouslyDefined(node, kind, field, (*val).Pos)
		}

		*val = boolFromValue(node, true)
		return nil
	}
}

func tokenFlagField(kind, field string, val **ast.Bool) func(*tokenNode) error {
	return func(node *tokenNode) error {
		if *val != nil {
			return previouslyDefined(node, kind, field, (*val).Pos)
		}

		*val = boolFromValue(node, true)
		return nil
	}
}

func lineAround(data []byte, offset int) (start, end int) {
	// find the index of the '\n' before data[offset]
	start = 0
	for i := offset - 1; i >= 0; i-- {
		if data[i] == '\n' {
			start = i + 1
			break
		}
	}

	// find the index of the '\n' after data[offset]
	end = len(data)
	for i := offset; i < len(data); i++ {
		if data[i] == '\n' {
			end = i
			break
		}
	}

	return start, end
}

func generateContext(source []byte, pos scanner.Position, length int) (
	context string) {

	var context_bytes []byte

	if pos.Offset > len(source) {
		panic("internal error: underline on strange position")
	}

	line_start, line_end := lineAround(source, pos.Offset)
	line := string(source[line_start:line_end])

	var before_line string
	if line_start > 0 {
		before_start, before_end := lineAround(source, line_start-1)
		before_line = string(source[before_start:before_end])
		before_line = strings.Replace(before_line, "\t", "    ", -1)
		context_bytes = append(context_bytes,
			fmt.Sprintf("% 4d: ", pos.Line-1)...)
		context_bytes = append(context_bytes, before_line...)
		context_bytes = append(context_bytes, '\n')
	}

	tabs := strings.Count(line, "\t")
	line = strings.Replace(line, "\t", "    ", -1)
	context_bytes = append(context_bytes, fmt.Sprintf("% 4d: ", pos.Line)...)
	context_bytes = append(context_bytes, line...)
	context_bytes = append(context_bytes, '\n')

	if length <= 0 {
		length = 1
	}
	offset := tabs*4 + pos.Column - 1 - tabs + 6
	underline := strings.Repeat(" ", offset) + strings.Repeat("^", length)
	context_bytes = append(context_bytes, underline...)

	return string(context_bytes)
}
