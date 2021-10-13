// Copyright 2021 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codec

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logging.TestingOverride()
}

type ObjectReader func() (interface{}, error)

func TestParseFromFile(t *testing.T) {
	tests := []struct {
		deepMindFile     string
		expectedPanicErr error
	}{
		{"testdata/deep-mind.dmlog", nil},
	}

	for _, test := range tests {
		t.Run(strings.Replace(test.deepMindFile, "testdata/", "", 1), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					require.Equal(t, test.expectedPanicErr, r)
				}
			}()

			cr := testFileConsoleReader(t, test.deepMindFile)
			buf := &bytes.Buffer{}
			buf.Write([]byte("["))

			for first := true; true; first = false {
				var reader ObjectReader = cr.Read
				out, err := reader()
				if v, ok := out.(proto.Message); ok && !isNil(v) {
					if !first {
						buf.Write([]byte(","))
					}

					// FIXMME: jsonpb needs to be updated to latest version of used gRPC
					//         elements. We are disaligned and using that breaks now.
					//         Needs to check what is the latest way to properly serialize
					//         Proto generated struct to JSON.
					// value, err := jsonpb.MarshalIndentToString(v, "  ")
					// require.NoError(t, err)

					value, err := json.MarshalIndent(v, "", "  ")
					require.NoError(t, err)

					buf.Write([]byte(value))
				}

				if err == io.EOF {
					break
				}

				if len(buf.Bytes()) != 0 {
					buf.Write([]byte("\n"))
				}

				require.NoError(t, err)
			}
			buf.Write([]byte("]"))

			goldenFile := test.deepMindFile + ".golden.json"
			if os.Getenv("GOLDEN_UPDATE") == "true" {
				ioutil.WriteFile(goldenFile, buf.Bytes(), os.ModePerm)
			}

			cnt, err := ioutil.ReadFile(goldenFile)
			require.NoError(t, err)

			if !assert.Equal(t, string(cnt), buf.String()) {
				t.Error("previous diff:\n" + unifiedDiff(t, cnt, buf.Bytes()))
			}
		})
	}
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

func testFileConsoleReader(t *testing.T, filename string) *ConsoleReader {
	t.Helper()

	fl, err := os.Open(filename)
	require.NoError(t, err)

	cr := testReaderConsoleReader(t, make(chan string, 10000), func() { fl.Close() })

	go cr.ProcessData(fl)

	return cr
}

func testReaderConsoleReader(t *testing.T, lines chan string, closer func()) *ConsoleReader {
	t.Helper()

	l := &ConsoleReader{
		lines: lines,
		close: closer,
		ctx:   &parseCtx{},
	}

	return l
}

func unifiedDiff(t *testing.T, cnt1, cnt2 []byte) string {
	file1 := "/tmp/gotests-linediff-1"
	file2 := "/tmp/gotests-linediff-2"
	err := ioutil.WriteFile(file1, cnt1, 0600)
	require.NoError(t, err)

	err = ioutil.WriteFile(file2, cnt2, 0600)
	require.NoError(t, err)

	cmd := exec.Command("diff", "-u", file1, file2)
	out, _ := cmd.Output()

	return string(out)
}
