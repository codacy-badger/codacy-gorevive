package main

import (
	toolparameters "codacy.com/codacy-gorevive/toolparameters"
	"fmt"
	codacy "github.com/codacy/codacy-engine-golang-seed"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	unnamedParamName     = "unnamedParam"
	sourceConfigFileName = "codacyrc.toml"
)

// paramValue converts codacy's parameter into a revive parameter value
func paramValue(param codacy.PatternParameter, patternID string) interface{} {
	ruleDefinition, notFound := toolparameters.FindRuleParameterDefinition(patternID)
	if notFound != nil {
		if isInteger(param.Value) {
			return int(param.Value.(float64))
		}
	}

	// check the type of parameter according to the tool documentation
	switch ruleDefinition.Type {
	case toolparameters.ListType:
		return strings.Split(param.Value.(string), ", ")
	case toolparameters.IntType:
		return int(param.Value.(float64))
	case toolparameters.FloatType:
		return param.Value.(float64)
	case toolparameters.StringType:
		return param.Value.(string)
	default:
		return param.Value
	}
}

func unnamedParam(value interface{}) []interface{} {
	resultTmp := []interface{}{}
	switch value.(type) {
	case []string:
		// if is a []string, append all values to res, one by one
		for _, v := range value.([]string) {
			resultTmp = append(resultTmp, v)
		}
	default:
		resultTmp = append(resultTmp, value)
	}
	return resultTmp
}

// patternParametersAsReviveValues converts pattern parameters into a list of revive arguments
func patternParametersAsReviveValues(pattern codacy.Pattern) []interface{} {
	if len(pattern.Parameters) == 0 {
		return []interface{}{}
	}

	namedParameters := map[string]interface{}{}
	for _, p := range pattern.Parameters {
		value := paramValue(p, pattern.PatternID)

		if p.Name == unnamedParamName {
			return unnamedParam(value)
		}

		namedParameters[p.Name] = value
	}

	if len(namedParameters) > 0 {
		return []interface{}{
			namedParameters,
		}
	}

	return []interface{}{}
}

func reviveArguments(paramsValues []interface{}) map[string]interface{} {
	if paramsValues == nil || len(paramsValues) == 0 {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"arguments": paramsValues,
	}
}

func reviveRuleName(id string) string {
	return "rule." + id
}

func patternsToReviveConfigMap(patterns []codacy.Pattern) map[string]interface{} {
	patternsMap := map[string]interface{}{}
	for _, pattern := range patterns {
		paramsValues := patternParametersAsReviveValues(pattern)
		patternsMap[reviveRuleName(pattern.PatternID)] = reviveArguments(paramsValues)
	}
	return patternsMap
}

func generateToolConfigurationContent(patterns []codacy.Pattern) string {
	patternsMap := patternsToReviveConfigMap(patterns)

	tomlString, err := mapToTOML(patternsMap)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return tomlString
}

func configurationFromSourceCode(sourceFolder string) (string, error) {
	filename := path.Join(sourceFolder, sourceConfigFileName)
	contentByte, err := ioutil.ReadFile(filename)
	return string(contentByte), err
}

// getConfigurationFile returns file, boolean saying if it is temp and error
func getConfigurationFile(patterns []codacy.Pattern, sourceFolder string) (*os.File, error) {
	// if no patterns, try to use configuration from source code
	// otherwise default configuration file
	if len(patterns) == 0 {
		sourceConfigFileContent, err := configurationFromSourceCode(sourceFolder)
		if err == nil {
			return writeToTempFile(sourceConfigFileContent)
		}

		return nil, err
	}

	content := generateToolConfigurationContent(patterns)

	return writeToTempFile(content)
}
