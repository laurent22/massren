package main

import (
	"path/filepath"
	"time"	
)

var historySize_ int = 500

type HistoryItem struct {
	Source string
	Dest string
	Timestamp int64
	Id string
}

func normalizePath(p string) string {
	output, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		panic(err)
	}
	return output
}

func saveHistoryItem(source string, dest string) error {
	_, err := profileDb_.Exec("INSERT INTO history (source, destination, timestamp) VALUES (?, ?, ?)", normalizePath(source), normalizePath(dest), time.Now().Unix())
	return err
}

func deleteHistoryItems(items []HistoryItem) error {
	if len(items) == 0 {
		return nil
	}
	
	sqlOr := ""
	for _, item := range items {
		if sqlOr != "" {
			sqlOr += " OR "
		}
		sqlOr += "id = " + item.Id
	}
	
	_, err := profileDb_.Exec("DELETE FROM history WHERE " + sqlOr)
	
	return err
}

// func latestHistoryItemsByDestinations(paths []string) []HistoryItem {
// 	var output []HistoryItem
	
// 	var pathParams []interface{}
// 	for _, p := range paths {
// 		pathParams = append(pathParams, p)
// 	}
	
// 	return output
// }

// func historyItems() ([]HistoryItem, error) {
// 	var output []HistoryItem
	
// 	// if _, err := os.Stat(historyFile()); os.IsNotExist(err) {
// 	// 	return output, nil
// 	// }
	
// 	// content, err := ioutil.ReadFile(historyFile())
// 	// if err != nil {
// 	// 	return output, err
// 	// }
	
// 	// lines := strings.Split(string(content), "\n")
// 	// for _, line := range lines {
// 	// 	line = strings.Trim(line, "\r\n\t ")
// 	// 	if line == "" {
// 	// 		continue
// 	// 	}
// 	// 	var item HistoryItem
// 	// 	json.Unmarshal([]byte(line), &item)
// 	// 	output = append(output, item)
// 	// }
	
// 	return output, nil
// }