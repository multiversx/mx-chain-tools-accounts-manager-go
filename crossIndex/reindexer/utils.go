package reindexer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
)

const (
	accountsTemplateFileName = "accounts.json"
	accountsPolicyFileName   = "accounts-policy.json"
	valuesIndex              = "values"
)

func readTemplateAndPolicyForAccountsIndex(pathToIndicesConfig string) (*bytes.Buffer, *bytes.Buffer, error) {
	pathTemplate := path.Join(pathToIndicesConfig, accountsTemplateFileName)
	template, err := readFile(pathTemplate)
	if err != nil {
		return nil, nil, err
	}

	pathPolicy := path.Join(pathToIndicesConfig, accountsPolicyFileName)
	policy, err := readFile(pathPolicy)
	if err != nil {
		return nil, nil, err
	}

	return template, policy, nil
}

func readTemplateForIndex(pathToIndicesConfig string, index string) (*bytes.Buffer, error) {
	templatePath := path.Join(pathToIndicesConfig, index) + ".json"
	return readFile(templatePath)
}

func readFile(path string) (*bytes.Buffer, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("readFile: %w, path %s, error %s", err, path, err.Error())
	}

	buff := &bytes.Buffer{}
	_, err = buff.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("readFile: %w, path %s, error %s", err, path, err.Error())
	}

	return buff, nil
}
