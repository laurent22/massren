package main

import (
	"os"	
)

func handleUndoCommand(opts *CommandLineOptions, args []string) error {
	filePaths, err := filePathsFromArgs(args)
	if err != nil {
		return err
	}
	
	items, err := historyItems()
	if err != nil {
		return err
	}
	
	var restoredItems []HistoryItem
	for _, filePath := range filePaths {
		filePath = normalizePath(filePath)
		for i := len(items) - 1; i >= 0; i-- {
			item := items[i]
			if filePath == item.Dest {
				if opts.DryRun {
					logInfo("\"%s\"  =>  \"%s\"", filePath, item.Source) 
				} else {
					logDebug("\"%s\"  =>  \"%s\"", filePath, item.Source) 
					err = os.Rename(filePath, item.Source)
					if err != nil {
						return err	
					}
					restoredItems = append(restoredItems, item)
				}
				break
			}
		}
	}
	
	deleteHistoryItems(restoredItems)

	return nil
}