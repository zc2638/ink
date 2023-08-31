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

package docker

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

func newDockerState() *dockerState {
	return &dockerState{
		volumes:    make(map[string]string),
		containers: make(map[string]string),
	}
}

type dockerState struct {
	id         string
	volumes    map[string]string
	containers map[string]string
}

type jsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *jsonError) Error() string {
	return e.Message
}

type jsonMessage struct {
	ID       string     `json:"id"`
	Status   string     `json:"status"`
	Error    *jsonError `json:"errorDetail"`
	Progress *struct{}  `json:"progressDetail"`
}

// PullReaderCopy copies a json message string to the io.Writer.
func PullReaderCopy(in io.Reader, out io.Writer) error {
	dec := json.NewDecoder(in)
	for {
		var jm jsonMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if jm.Error != nil {
			if jm.Error.Code == 401 {
				return fmt.Errorf("authentication is required")
			}
			return jm.Error
		}

		if jm.Progress != nil {
			continue
		}
		if jm.ID == "" {
			fmt.Fprintf(out, "%s\n", jm.Status)
		} else {
			fmt.Fprintf(out, "%s: %s\n", jm.ID, jm.Status)
		}
	}
	return nil
}

type StdType byte

const (
	Stdin  StdType = iota // Stdin represents the standard input stream type.
	Stdout                // Stdout represents the standard output stream type.
	Stderr                // Stderr represents the standard error steam type.
)

const (
	stdWriterPrefixLen = 8
	stdWriterFdIndex   = 0
	stdWriterSizeIndex = 4

	startingBufLen = 32*1024 + stdWriterPrefixLen + 1
)

func StdCopy(dist io.Writer, src io.Reader) (written int64, err error) {
	var (
		buf       = make([]byte, startingBufLen)
		bufLen    = len(buf)
		nr, nw    int
		er, ew    error
		out       io.Writer
		frameSize int
	)

	for {
		// Make sure we have at least a full header
		for nr < stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		// Check the first byte to know where to write
		switch StdType(buf[stdWriterFdIndex]) {
		case Stdin:
			fallthrough
		case Stdout:
			// Write on stdout
			out = dist
		case Stderr:
			// Write on stderr
			out = dist
		default:
			return 0, fmt.Errorf("unrecognized input header: %d", buf[stdWriterFdIndex])
		}

		// Retrieve the size of the frame
		frameSize = int(binary.BigEndian.Uint32(buf[stdWriterSizeIndex : stdWriterSizeIndex+4]))

		// Check if the buffer is big enough to read the frame.
		// Extend it if necessary.
		if frameSize+stdWriterPrefixLen > bufLen {
			buf = append(buf, make([]byte, frameSize+stdWriterPrefixLen-bufLen+1)...)
			bufLen = len(buf)
		}

		// While the number of bytes read is less than the size of the frame + header, we keep reading
		for nr < frameSize+stdWriterPrefixLen {
			var nr2 int
			nr2, er = src.Read(buf[nr:])
			nr += nr2
			if er == io.EOF {
				if nr < frameSize+stdWriterPrefixLen {
					return written, nil
				}
				break
			}
			if er != nil {
				return 0, er
			}
		}

		// Write the retrieved frame (without a header)
		nw, ew = out.Write(buf[stdWriterPrefixLen : frameSize+stdWriterPrefixLen])
		if ew != nil {
			return 0, ew
		}
		// If the frame has not been fully written: error
		if nw != frameSize {
			return 0, io.ErrShortWrite
		}
		written += int64(nw)

		// Move the rest of the buffer to the beginning
		copy(buf, buf[frameSize+stdWriterPrefixLen:])
		// Move the index
		nr -= frameSize + stdWriterPrefixLen
	}
}
