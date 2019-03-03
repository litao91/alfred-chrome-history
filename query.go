package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pasela/alfred-chrome-history/history"
	"github.com/pasela/alfred-chrome-history/utils"
)

func queryHistory(profile, url, title string, limit int) ([]history.Entry, error) {
	histFile := historyFile{
		Profile: profile,
		Clone:   false,
	}
	defer histFile.Close()

	filePath, err := histFile.GetPath()
	if err != nil {
		return nil, err
	}

	his, err := history.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer his.Close()

	return his.Query(url, title, limit)
}

type historyFile struct {
	Profile string
	Clone   bool

	tempDir    string
	autoRemove bool
}

func (h *historyFile) initTempDir() error {
	if !h.Clone || h.tempDir != "" {
		return nil
	}

	tempDir, err := ioutil.TempDir("", "alfred-chrome-history")
	if err != nil {
		return err
	}
	h.tempDir = tempDir
	h.autoRemove = true
	return nil
}

func (h *historyFile) Close() error {
	if h.autoRemove && h.tempDir != "" {
		return os.RemoveAll(h.tempDir)
	}
	return nil
}

func (h *historyFile) GetPath() (string, error) {
	origFile, err := history.GetHistoryPath(h.Profile)
	if err != nil {
		return "", err
	}
	if !h.Clone {
		return origFile, nil
	}

	if err := h.initTempDir(); err != nil {
		return "", err
	}

	clonedFile := filepath.Join(h.tempDir, filepath.Base(origFile))
	if _, err := utils.CopyFile(origFile, clonedFile); err != nil {
		return "", err
	}
	return clonedFile, nil
}

func run() error {
	var limit int
	profile := os.Getenv("CHROME_PROFILE")
	flag.StringVar(&profile, "profile", profile, "Chrome profile directory")
	flag.IntVar(&limit, "limit", 0, "Limit n results")
	flag.Parse()

	query := strings.Join(flag.Args(), " ")
	entries, err := queryHistory(profile, query, query, limit)
	if err != nil {
		return err
	}

	r := make([]map[string]string, 0, 0)

	for _, entry := range entries {
		i := make(map[string]string)
		i["text"] = entry.Title
		i["subtext"] = entry.URL
		i["arg"] = entry.URL
		i["plugin"] = "seal_chistory"
		r = append(r, i)
	}
	str, _ := json.Marshal(r)
	fmt.Println(string(str))

	return nil
}

func main() {
	run()
}
