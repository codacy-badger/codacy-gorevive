package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func downloadFile(url string) (*os.File, error) {
	out, err := ioutil.TempFile(os.TempDir(), "tmp-gorevive-")
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func getRulesDescriptionMarkdown(toolVersion string) (*os.File, error) {
	rulesDescriptionFileURL := fmt.Sprintf("https://raw.githubusercontent.com/mgechev/revive/v%s/RULES_DESCRIPTIONS.md", toolVersion)
	return downloadFile(rulesDescriptionFileURL)
}
