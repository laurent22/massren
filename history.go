package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"	
)

type HistoryItem struct {
	Source string
	Dest string
	Timestamp int64
}

func historyFile() string {
	return configFolder() + "/history"
}

func saveHistory(source string, dest string) error {
	f, err := os.OpenFile(historyFile(), os.O_APPEND | os.O_CREATE | os.O_WRONLY, CONFIG_PERM)
	if err != nil {
		return err
	}
	defer f.Close()

	item := HistoryItem{
		Source: source,
		Dest: dest,
		Timestamp: time.Now().Unix(),
	}
	
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

func history() ([]HistoryItem, error) {
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
		var item HistoryItem
		json.Unmarshal([]byte(line), &item)
		output = append(output, item)
	}
	
	return output, nil
}