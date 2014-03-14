package main

import (
	"path/filepath"
	"time"
)

type HistoryItem struct {
	Source    string
	Dest      string
	Timestamp int64
	Id        string
}

func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	output, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		panic(err)
	}
	return output
}

func clearHistory() error {
	_, err := profileDb_.Exec("DELETE FROM history")
	return err
}

func allHistoryItems() ([]HistoryItem, error) {
	var output []HistoryItem

	rows, err := profileDb_.Query("SELECT id, source, destination, timestamp FROM history ORDER BY id")
	if err != nil {
		return output, err
	}

	for rows.Next() {
		var item HistoryItem
		rows.Scan(&item.Id, &item.Source, &item.Dest, &item.Timestamp)
		output = append(output, item)
	}

	return output, nil
}

func saveHistoryItems(fileActions []*FileAction) error {
	if len(fileActions) == 0 {
		return nil
	}

	tx, err := profileDb_.Begin()
	if err != nil {
		return err
	}

	for _, action := range fileActions {
		if action.kind == KIND_DELETE {
			// Current, undo is not supported
			continue
		}
		tx.Exec("INSERT INTO history (source, destination, timestamp) VALUES (?, ?, ?)", normalizePath(action.oldPath), normalizePath(action.newPath), time.Now().Unix())
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
	if profileDb_ != nil {
		profileDb_.Exec("DELETE FROM history WHERE timestamp < ?", minTimestamp)
	}
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

	rows, err := profileDb_.Query("SELECT id, source, destination, timestamp FROM history WHERE "+sqlOr+" ORDER BY timestamp DESC", sqlArgs...)
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
