package main

import (
	"fmt"
	codacy "github.com/codacy/codacy-golang-tools-engine"
	"io/ioutil"
	"os"
	"path"
)

var unnamedParamName = "unnamedParam"

func patternToToml(pattern codacy.Pattern) string {
	ruleString := "[rule." + pattern.PatternID + "]"

	patternParams := ""
	if len(pattern.Parameters) == 1 && pattern.Parameters[0].Name == unnamedParamName {
		patternParams = "arguments = [" + fmt.Sprintf("%v", pattern.Parameters[0].Value) + "]"
	} else {
		params := "{"
		for _, param := range pattern.Parameters {
			if params != "{" {
				params = params + ","
			}
			params = params + fmt.Sprintf(`%s = "%s"`, param.Name, param.Value)
		}
		params = params + "}"
		patternParams = "arguments = [" + params + "]"
	}

	return fmt.Sprintf("%s\n  %s", ruleString, patternParams)
}

func generateToolConfigurationContent(patterns []codacy.Pattern) string {
	content := ""
	for _, p := range patterns {
		patternsTomlString := patternToToml(p)
		content = fmt.Sprintf("%s%s\n\n", content, patternsTomlString)
	}
	return content
}

func writeConfigurationToTempFile(content string) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "gorevive-")
	if err != nil {
		return nil, err
	}
	if _, err = tmpFile.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func getConfigurationFromSourceCode(sourceFolder string) (string, error) {
	filename := path.Join(sourceFolder, "codacyrc.toml")

	contentByte, err := ioutil.ReadFile(filename)
	return string(contentByte), err
}

// getConfigurationFile returns file, boolean saying if it is temp and error
func getConfigurationFile(patterns []codacy.Pattern, sourceFolder string) (*os.File, error) {
	// if no patterns, try to use configuration from source code
	// otherwise default configuration file
	if len(patterns) == 0 {
		sourceConfigFileContent, err := getConfigurationFromSourceCode(sourceFolder)
		if err == nil {
			return writeConfigurationToTempFile(sourceConfigFileContent)
		}

		return nil, nil
	}

	content := generateToolConfigurationContent(patterns)

	return writeConfigurationToTempFile(content)
}
