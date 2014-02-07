package main

import (
	"path/filepath"
	"testing"
)

func Test_saveHistoryItems(t *testing.T) {
	setup(t)
	defer teardown(t)
	
	err := saveHistoryItems([]string{"one", "two"}, []string{"one"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	
	err = saveHistoryItems([]string{}, []string{})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	
	items, _ := allHistoryItems()
	if len(items) > 0 {
		t.Errorf("Expected no items, got %d", len(items))
	}
	
	saveHistoryItems([]string{"one","two"}, []string{"1","2"})
	
	items, _ = allHistoryItems()
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	
	for _, item := range items {
		if (filepath.Base(item.Source) == "one" && filepath.Base(item.Dest) != "1") || (filepath.Base(item.Source) == "two" && filepath.Base(item.Dest) != "2") {
			t.Error("Source and destination do not match.")
		}
	}
	
	saveHistoryItems([]string{"three"}, []string{"3"})
	items, _ = allHistoryItems()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
	
	profileDb_.Close()
		
	err = saveHistoryItems([]string{"un"}, []string{"dest"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func Test_deleteHistoryItems(t *testing.T) {
	setup(t)
	defer teardown(t)

	saveHistoryItems([]string{"one","two","three"}, []string{"1","2","3"})
	
	items, _ := allHistoryItems()
	deleteHistoryItems([]HistoryItem{items[0], items[1]})
	
	items, _ = allHistoryItems()
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	} else {
		if filepath.Base(items[0].Source) != "three" {
			t.Error("Incorrect item in history")
		}
	}
}

func Test_deleteOldHistoryItems(t *testing.T) {
	setup(t)
	defer teardown(t)

	now := 1000
	for i := 0; i < 5; i++ {
		profileDb_.Exec("INSERT INTO history (source, destination, timestamp) VALUES (?, ?, ?)", "a", "b", now + i)	
	}
	deleteOldHistoryItems(int64(now + 2))
	
	items, _ := allHistoryItems()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func Test_latestHistoryItemsByDestinations(t *testing.T) {
	setup(t)
	defer teardown(t)

	now := 1000
	for i := 0; i < 5; i++ {
		profileDb_.Exec("INSERT INTO history (source, destination, timestamp) VALUES (?, ?, ?)", "a", "b", now + i)	
	}
	
	items, _ := allHistoryItems()
	dest := items[0].Dest
	
	items, _ = latestHistoryItemsByDestinations([]string{dest})
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	} else {
		if items[0].Timestamp != 1004 {
			t.Error("Did not get the right item")			
		}
	}
}