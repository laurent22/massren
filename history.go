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

func latestHistoryItemsByDestinations(paths []string) ([]HistoryItem, error) {	
	var output []HistoryItem
	if len(paths) == 0 {
		return output, nil
	}
	
	sqlOr := "" 
	var sqlArgs []interface{}
	for _, p := range paths {
		sqlArgs = append(sqlArgs, p)
		if sqlOr != "" {
			sqlOr += " OR "
		}
		sqlOr += "destination = ?"
	}
	
	rows, err := profileDb_.Query("SELECT id, source, destination, timestamp FROM history WHERE " + sqlOr + " ORDER BY timestamp DESC", sqlArgs...)
	if err != nil {
		return output, err
	}
	
	doneDestinations := make(map[string]bool)
	for rows.Next() {
		var item HistoryItem
		rows.Scan(&item.Id, &item.Source, &item.Dest, &item.Timestamp)
		_, done := doneDestinations[item.Dest]
		if done {
			continue
		}
		output = append(output, item)
		doneDestinations[item.Dest] = true
	}
	
	return output, nil
}