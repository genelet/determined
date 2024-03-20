package dethcl

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type Slack struct {
	Channel string `hcl:"channel"`
	Message string `hcl:"message"`
}

type Python struct {
	PythonName    string `hcl:"python_name,label"`
	PythonVersion int    `hcl:"python_version,optional"`
	Path          string `hcl:"root_dir,optional"`
}

type Job struct {
	JobName       string  `hcl:"job_name,label"`
	Description   string  `hcl:"description,label"`
	ProgramPython *Python `hcl:"python,block"`
	ProgramSlack  *Slack  `hcl:"slack,block"`
}

type Pipeline struct {
	Version      int               `hcl:"version,optional"`
	Say          map[string]string `hcl:"say,optional"`
	TEST_FOLDER  string            `hcl:"TEST_FOLDER"`
	EXECUTION_ID string            `hcl:"EXECUTION_ID"`
	Jobs         []*Job            `hcl:"job,block"`
}

func TestDynaCty(t *testing.T) {
	data1 := `
TEST_FOLDER = "__test__"
EXECUTION_ID = random(6)
version = 2
say = {
	for k, v in {hello: "world"}: k => v if k == "hello"
}

job check "this is a temporal job" {
  python "run.py" {}
}

job e2e "running integration tests" {

  python "app-e2e.py" {
    root_dir = var.TEST_FOLDER
	python_version = version + 6
  }

  slack {
    channel  = "slack-my-channel"
    message = "Job execution ${EXECUTION_ID} completed successfully"
  }
}
`
	p := new(Pipeline)
	ref := map[string]interface{}{
		"functions": map[string]function.Function{
			"random": function.New(&function.Spec{

				VarParam: nil,
				Params: []function.Parameter{
					{Type: cty.Number},
				},
				Type: func(args []cty.Value) (cty.Type, error) {
					return cty.String, nil
				},
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
					n, _ := args[0].AsBigFloat().Int64()
					b := make([]rune, n)
					for i := range b {
						b[i] = letterRunes[rand.Intn(len(letterRunes))]
					}
					return cty.StringVal(string(b)), nil
				},
			})}}

	err := UnmarshalSpec([]byte(data1), p, nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if p.Say["hello"] != "world" {
		t.Errorf("%#v", p.Say)
	}
	if p.TEST_FOLDER != "__test__" || len(p.Jobs) != 2 {
		t.Errorf("%#v", p)
	}
	for i, job := range p.Jobs {
		if i == 0 {
			if job.JobName != "check" || job.ProgramPython.PythonName != "run.py" {
				t.Errorf("%#v", job)
			}
		}
		if i == 1 {
			if job.Description != "running integration tests" || job.ProgramPython.Path != "__test__" {
				t.Errorf("%#v", job)
			}
			python := job.ProgramPython
			if python.PythonVersion != 8 {
				t.Errorf("%#v", python)
			}
			slack := job.ProgramSlack
			if len(slack.Message) != 43 {
				t.Errorf("%#v", slack)
			}
		}
	}
}

func TestDynaNorm(t *testing.T) {
	data1 := `
TEST_FOLDER = "__test__"
EXECUTION_ID = random(6)
version = 2
say = {
	for k, v in {hello: "world"}: k => v if k == "hello"
}

job check "this is a temporal job" {
  python "run.py" {}
}

job e2e "running integration tests" {

  python "app-e2e.py" {
    root_dir = var.TEST_FOLDER
	python_version = version + 6
  }

  slack {
    channel  = "slack-my-channel"
    message = "Job execution ${EXECUTION_ID} completed successfully"
  }
}
`
	p := new(Pipeline)
	ref := map[string]interface{}{
		"functions": map[string]interface{}{
			"random": func(n int) string {
				var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
				b := make([]rune, n)
				for i := range b {
					b[i] = letterRunes[rand.Intn(len(letterRunes))]
				}
				return string(b)
			},
		},
	}

	err := UnmarshalSpec([]byte(data1), p, nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if p.Say["hello"] != "world" {
		t.Errorf("%#v", p.Say)
	}
	if p.TEST_FOLDER != "__test__" || len(p.Jobs) != 2 {
		t.Errorf("%#v", p)
	}
	for i, job := range p.Jobs {
		if i == 0 {
			if job.JobName != "check" || job.ProgramPython.PythonName != "run.py" {
				t.Errorf("%#v", job)
			}
		}
		if i == 1 {
			if job.Description != "running integration tests" || job.ProgramPython.Path != "__test__" {
				t.Errorf("%#v", job)
			}
			python := job.ProgramPython
			if python.PythonVersion != 8 {
				t.Errorf("%#v", python)
			}
			slack := job.ProgramSlack
			if len(slack.Message) != 43 {
				t.Errorf("%#v", slack)
			}
		}
	}
}

func TestDynaTime(t *testing.T) {
	data1 := `
TEST_FOLDER = "__test__"
EXECUTION_ID = datetimeparse("Jan 2, 2006 at 3:04pm (MST)", "Feb 3, 2013 at 7:54pm (PST)")
version = 2
say = {
	for k, v in {hello: "world"}: k => v if k == "hello"
}

job check "this is a temporal job" {
  python "run.py" {}
}

job e2e "running integration tests" {

  python "app-e2e.py" {
    root_dir = var.TEST_FOLDER
	python_version = version + 6
  }

  slack {
    channel  = "slack-my-channel"
    message = "${EXECUTION_ID}"
  }
}
`
	p := new(Pipeline)
	ref := map[string]interface{}{
		"functions": map[string]interface{}{
			"datetimeparse": func(layout, value string) (int64, error) {
				t, err := time.Parse(layout, value)
				if err != nil {
					return 0, err
				}
				return t.Unix(), nil
			},
		},
	}

	err := UnmarshalSpec([]byte(data1), p, nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if p.Say["hello"] != "world" {
		t.Errorf("%#v", p.Say)
	}
	if p.TEST_FOLDER != "__test__" || len(p.Jobs) != 2 {
		t.Errorf("%#v", p)
	}
	for i, job := range p.Jobs {
		if i == 0 {
			if job.JobName != "check" || job.ProgramPython.PythonName != "run.py" {
				t.Errorf("%#v", job)
			}
		}
		if i == 1 {
			if job.Description != "running integration tests" || job.ProgramPython.Path != "__test__" {
				t.Errorf("%#v", job)
			}
			python := job.ProgramPython
			if python.PythonVersion != 8 {
				t.Errorf("%#v", python)
			}
			slack := job.ProgramSlack
			for _, v := range strings.SplitN(slack.Message, "", -1) {
				if v != "1" && v != "3" && v != "8" && v != "9" && v != "7" && v != "5" && v != "4" && v != "2" && v != "0" && v != "6" {
					t.Errorf("%#v", slack.Message)
				}
			}
		}
	}
}
