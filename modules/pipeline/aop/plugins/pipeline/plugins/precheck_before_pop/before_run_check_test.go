// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package precheck_before_pop

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/pipeline_network_hook_client"
	"github.com/stretchr/testify/assert"
)

func Test_matchHookType(t *testing.T) {
	var table = []struct {
		lifecycle []*pipelineyml.NetworkHookInfo
		matchLen  int
	}{
		{
			lifecycle: []*pipelineyml.NetworkHookInfo{
				{
					Hook: HookType,
				},
				{
					Hook: "after-run-check",
				},
			},
			matchLen: 1,
		},

		{
			lifecycle: []*pipelineyml.NetworkHookInfo{
				{
					Hook: "pre-run-check",
				},
				{
					Hook: "after-run-check",
				},
			},
			matchLen: 0,
		},

		{
			lifecycle: []*pipelineyml.NetworkHookInfo{
				{
					Hook: HookType,
				},
				{
					Hook: "after-create-check",
				},
				{
					Hook: HookType,
				},
			},
			matchLen: 2,
		},
	}
	var httpBeforeCheckRun HttpBeforeCheckRun
	for _, data := range table {
		result := httpBeforeCheckRun.matchHookType(data.lifecycle)
		assert.Len(t, result, data.matchLen)
	}
}

func TestCheckRun(t *testing.T) {
	var table = []struct {
		CheckResult           CheckRunResult
		haveError             bool
		matchOtherLabel       string
		httpBeforeCheckRun    HttpBeforeCheckRun
		mockPipelineWithTasks *spec.PipelineWithTasks
	}{
		{
			CheckResult: CheckRunResult{
				CheckResult: CheckResultSuccess,
			},
			matchOtherLabel: "otherLabel",
			haveError:       false,
			httpBeforeCheckRun: HttpBeforeCheckRun{
				PipelineID: 1000,
			},
			mockPipelineWithTasks: &spec.PipelineWithTasks{
				Pipeline: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID:              1000,
						PipelineSource:  "FDP",
						PipelineYmlName: "230",
					},
					PipelineExtra: spec.PipelineExtra{
						PipelineYml: `
version: 1.1
lifecycle:
  - hook: "` + HookType + `"
    labels: 
       "otherLabel": 123
`,
					},
				},
				Tasks: []*spec.PipelineTask{},
			},
		},
		{
			CheckResult: CheckRunResult{
				CheckResult: CheckResultSuccess,
			},
			haveError:       false,
			matchOtherLabel: "otherLabel2",
			httpBeforeCheckRun: HttpBeforeCheckRun{
				PipelineID: 1000,
			},
			mockPipelineWithTasks: &spec.PipelineWithTasks{
				Pipeline: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID:              1000,
						PipelineSource:  "FDP",
						PipelineYmlName: "230",
					},
					PipelineExtra: spec.PipelineExtra{
						PipelineYml: `
version: 1.1
lifecycle:
  - hook: "` + HookType + `"
    labels: 
       "otherLabel2": 123
  - hook: "` + HookType + `"
    labels: 
       "otherLabel2": 123
`,
					},
				},
				Tasks: []*spec.PipelineTask{},
			},
		},
		{
			CheckResult: CheckRunResult{
				CheckResult: CheckResultSuccess,
			},
			haveError:       false,
			matchOtherLabel: "",
			httpBeforeCheckRun: HttpBeforeCheckRun{
				PipelineID: 10001,
			},
			mockPipelineWithTasks: &spec.PipelineWithTasks{
				Pipeline: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID:              10001,
						PipelineSource:  "FDP",
						PipelineYmlName: "230",
					},
					PipelineExtra: spec.PipelineExtra{
						PipelineYml: `
version: 1.1
lifecycle:
  - hook: "` + HookType + `1"
    labels: 
       "otherLabel1": 123
`,
					},
				},
				Tasks: []*spec.PipelineTask{},
			},
		},

		{
			CheckResult: CheckRunResult{
				CheckResult: CheckResultFailed,
				RetryOption: RetryOption{
					IntervalSecond: 203,
				},
			},
			haveError:       false,
			matchOtherLabel: "otherLabel1",
			httpBeforeCheckRun: HttpBeforeCheckRun{
				PipelineID: 10001,
			},
			mockPipelineWithTasks: &spec.PipelineWithTasks{
				Pipeline: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						ID:              10001,
						PipelineSource:  "FDP",
						PipelineYmlName: "230",
					},
					PipelineExtra: spec.PipelineExtra{
						PipelineYml: `
version: 1.1
lifecycle:
  - hook: "` + HookType + `"
    labels: 
       "otherLabel1": 123
`,
					},
				},
				Tasks: []*spec.PipelineTask{},
			},
		},

		{
			CheckResult: CheckRunResult{
				CheckResult: CheckResultFailed,
				RetryOption: RetryOption{
					IntervalSecond: 203,
				},
			},
			haveError:       true,
			matchOtherLabel: "otherLabel1",
			httpBeforeCheckRun: HttpBeforeCheckRun{
				PipelineID: 0,
			},
		},
	}

	for _, v := range table {
		var e dbclient.Client
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(&e), "GetPipelineWithTasks", func(client *dbclient.Client, id uint64) (*spec.PipelineWithTasks, error) {
			return v.mockPipelineWithTasks, nil
		})
		guard1 := monkey.PatchInstanceMethod(reflect.TypeOf(&e), "ListLabelsByPipelineID", func(client *dbclient.Client, pipelineID uint64, ops ...dbclient.SessionOption) ([]spec.PipelineLabel, error) {
			return nil, nil
		})
		guard2 := monkey.Patch(pipeline_network_hook_client.PostLifecycleHookHttpClient, func(source string, req interface{}, resp interface{}) error {
			checkRunResultRequest := req.(CheckRunResultRequest)
			if checkRunResultRequest.Labels != nil {
				pipelineLabels := checkRunResultRequest.Labels["pipelineLabels"].(map[string]interface{})
				assert.Equal(t, pipelineLabels[v.matchOtherLabel], 123)
			}

			var checkRunResultResponse CheckRunResultResponse
			checkRunResultResponse.Success = true
			checkRunResultResponse.CheckRunResult = v.CheckResult

			checkRunResultResponseJson, _ := json.Marshal(checkRunResultResponse)
			buffer := bytes.NewBuffer(checkRunResultResponseJson)
			err := json.NewDecoder(buffer).Decode(&resp)
			assert.NoError(t, err)
			return nil
		})
		v.httpBeforeCheckRun.DBClient = &e
		defer guard.Unpatch()
		defer guard1.Unpatch()
		defer guard2.Unpatch()

		result, err := v.httpBeforeCheckRun.CheckRun()
		if err != nil {
			assert.True(t, v.haveError, err)
		} else {
			assert.NotEmpty(t, result)
			assert.Equal(t, v.CheckResult.CheckResult, result.CheckResult)
			assert.Equal(t, v.CheckResult.RetryOption.IntervalSecond, result.RetryOption.IntervalSecond)
		}
	}

}
