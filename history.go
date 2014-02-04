package main

import (
	"errors"
	"path/filepath"
	"time"	
)

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

func saveHistoryItems(sources []string, destinations []string) error {
	if len(sources) != len(destinations) {
		return errors.New("Number of sources and destinations do not match.")
	}
	
	if len(sources) == 0 {
		return nil
	}
	
	tx, err := profileDb_.Begin()
	if err != nil {
		return err
	}
	
	for i, source := range sources {
		dest := destinations[i]
		profileDb_.Exec("INSERT INTO history (source, destination, timestamp) VALUES (?, ?, ?)", normalizePath(source), normalizePath(dest), time.Now().Unix())	
	}
	
	return tx.Commit()
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

func deleteOldHistoryItems(minTimestamp int64) {
	profileDb_.Exec("DELETE FROM history WHERE timestamp < ?", minTimestamp)
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