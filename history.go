package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"	
)

type HistoryItem struct {
	Source string
	Dest string
	Timestamp int64
	Id string
}

func historyFile() string {
	return configFolder() + "/history"
}

func normalizePath(p string) string {
	output, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		panic(err)
	}
	return output
}

func saveHistoryItem(source string, dest string) error {
	f, err := os.OpenFile(historyFile(), os.O_APPEND | os.O_CREATE | os.O_WRONLY, CONFIG_PERM)
	if err != nil {
		return err
	}
	defer f.Close()

	item := HistoryItem{
		Source: normalizePath(source),
		Dest: normalizePath(dest),
		Timestamp: time.Now().Unix(),
	}
	
	item.Id = stringHash(item.Source + "|" + item.Dest)
	
	b, err := json.Marshal(item)
	if err != nil {
		return err
	}
	
	_, err = f.WriteString(string(b) + "\n")
	if err != nil {
		return err
	}
	
	return nil
}

func deleteHistoryItems(items []HistoryItem) error {
	currentItems, err := historyItems()
	if err != nil {
		return err
	}
	var newItems []HistoryItem

	for _, item1 := range currentItems {
		found := false
		for _, item2 := range items {
			if item1.Id == item2.Id {
				found = true
				break
			}
		}
		if !found {
			newItems = append(newItems, item1)
		}
	}
	
	f, err := os.OpenFile(historyFile(), os.O_TRUNC | os.O_CREATE | os.O_RDWR, CONFIG_PERM)
	if err != nil {
		return err
	}
	defer f.Close()
	
	for _, item := range newItems {
		b, err := json.Marshal(item)
		if err != nil {
			return err
		}
		_, err = f.WriteString(string(b) + "\n")
		if err != nil {
			return err
		}
	}
	
	return nil
}

func historyItems() ([]HistoryItem, error) {
	var output []HistoryItem
	
	if _, err := os.Stat(historyFile()); os.IsNotExist(err) {
		return output, nil
	}
	
	content, err := ioutil.ReadFile(historyFile())
	if err != nil {
		return output, err
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.Trim(line, "\r\n\t ")
		if line == "" {
			continue
		}
		var item HistoryItem
		json.Unmarshal([]byte(line), &item)
		output = append(output, item)
	}
	
	return output, nil
}