// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	testcommon "github.com/maximilien/kn-source-pkg/test/e2e"
	"gotest.tools/v3/assert"
	"knative.dev/client/lib/test"
)

const pluginName string = "admin"

type e2eTest struct {
	it         *testcommon.E2ETest
	kn         *test.Kn
	kubectl    *test.Kubectl
	backupData map[string]string
}

func newE2ETest(t *testing.T) *e2eTest {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil
	}

	it, err := testcommon.NewE2ETest("kn-diag", filepath.Join(currentDir, "../.."), false)
	if err != nil {
		return nil
	}

	kn := test.NewKn()
	kubectl := test.NewKubectl("knative-serving")
	e2eTest := &e2eTest{
		it:         it,
		kn:         &kn,
		kubectl:    &kubectl,
		backupData: make(map[string]string),
	}
	return e2eTest
}

func TestKnDiagPlugin(t *testing.T) {
	t.Parallel()

	e2eTest := newE2ETest(t)
	assert.Assert(t, e2eTest != nil)
	defer func() {
		assert.NilError(t, e2eTest.it.KnTest().Teardown())
	}()

	r := test.NewKnRunResultCollector(t, e2eTest.it.KnTest())
	defer r.DumpIfFailed()

	err := e2eTest.it.KnPlugin().Install()
	assert.NilError(t, err)

	err = e2eTest.it.KnPlugin().Uninstall()
	assert.NilError(t, err)
}
