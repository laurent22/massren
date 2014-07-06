package main

import (
	"github.com/nu7hatch/gouuid"
	"os"
)

func handleUndoCommand(opts *CommandLineOptions, args []string) error {
	filePaths, err := filePathsFromArgs(args)
	if err != nil {
		return err
	}

	for i, p := range filePaths {
		filePaths[i] = normalizePath(p)
	}

	items, err := latestHistoryItemsByDestinations(filePaths)
	if err != nil {
		return err
	}

	var conflictItems []HistoryItem

	for _, item := range items {
		if opts.DryRun {
			logInfo("\"%s\"  =>  \"%s\"", item.Dest, item.Source)
		} else {
			logDebug("\"%s\"  =>  \"%s\"", item.Dest, item.Source)

			if _, err := os.Stat(item.Source); os.IsNotExist(err) {
				err = os.Rename(item.Dest, item.Source)
				if err != nil {
					return err
				}
			} else {
				u, _ := uuid.NewV4()
				item.IntermediatePath = item.Source + "-" + u.String()
				conflictItems = append(conflictItems, item)
			}
		}
	}

	// See conflict resolution in main::processFileActions()

	for _, item := range conflictItems {
		err := os.Rename(item.Dest, item.IntermediatePath)
		if err != nil {
			return err
		}
	}

	for _, item := range conflictItems {
		err := os.Rename(item.IntermediatePath, item.Source)
		if err != nil {
			return err
		}
	}

	deleteHistoryItems(items)

	return nil
}
