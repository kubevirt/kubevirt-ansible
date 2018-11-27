package framework

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/rand"
)

type TestRandom struct {
	testDir         string
	generateName    string
	generateABSPath string
}

const TestDir = "kubevirt-ansible-test"

func NewTestRandom() (*TestRandom, error) {
	var err error
	t := new(TestRandom)
	t.testDir, err = ioutil.TempDir("", TestDir)
	if err != nil {
		return nil, err
	}
	t.generateName = "generate-" + rand.String(10)
	t.generateABSPath = filepath.Join(t.testDir, t.generateName+".json")
	return t, nil
}

func (t *TestRandom) Name() string {
	return t.generateName
}

func (t *TestRandom) ABSPath() string {
	return t.generateABSPath
}

func (t *TestRandom) CleanUp() error {
	if t.testDir != "" {
		return os.RemoveAll(t.testDir)
	} else {
		return fmt.Errorf("testDir is empty")
	}
}
