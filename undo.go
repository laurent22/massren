package main

import (
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

	for _, item := range items {
		if opts.DryRun {
			logInfo("\"%s\"  =>  \"%s\"", item.Dest, item.Source) 
		} else {
			logDebug("\"%s\"  =>  \"%s\"", item.Dest, item.Source) 
			err = os.Rename(item.Dest, item.Source)
			if err != nil {
				return err	
			}
		}
	}
	
	deleteHistoryItems(items)

	return nil
}