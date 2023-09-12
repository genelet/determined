package dethcl

import (
	"fmt"
	"math/rand"
	"testing"
//	"github.com/zclconf/go-cty/cty"
)

type Slack struct {
	Channel string `hcl:"channel"`
	Message string `hcl:"message"`
}

type Python struct {
	PythonName string `hcl:"python_name,label"`
    Path string `hcl:"root_dir,optional"`
}

type Job struct {
	JobName string `hcl:"job_name,label"`
	Description string `hcl:"description,label"`
	ProgramPython *Python `hcl:"python,block"`
	ProgramSlack *Slack `hcl:"slack,block"`
}

type Pipeline struct {
    Version int `hcl:"version,optional"`
    TEST_FOLDER string `hcl:"TEST_FOLDER"`
    EXECUTION_ID string `hcl:"EXECUTION_ID"`
	Jobs []*Job  `hcl:"job,block"`
}

func TestDyna(t *testing.T) {
	data1 := `
version = 2
TEST_FOLDER = "__test__"
EXECUTION_ID = random(6)

job check "this is a temporal job" {
  python "run.py" {}
}

job e2e "running integration tests" {

  python "app-e2e.py" {
    root_dir = var.TEST_FOLDER
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
			"random": func(num ...interface{}) (interface{}, error) {
				n := 6
				if len(num) == 1 {
					switch t := num[0].(type) {
					case float64:
						n = int(t + 0.000001)
					default:
						n = num[0].(int)
					}
				} else if len(num) > 1 {
					return nil, fmt.Errorf("wrong args size %d in function random", len(num))
				}
				var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
        		b := make([]rune, n)
        		for i := range b {
            		b[i] = letterRunes[rand.Intn(len(letterRunes))]
        		}
        		return string(b), nil
			},
		},
	}

	err := UnmarshalSpec([]byte(data1), p, nil, ref)
	if err != nil {
		t.Fatal(err)
	}
	if p.TEST_FOLDER != "__test__"  || len(p.Jobs) != 2 {
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
		}
	}
}
