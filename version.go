package main

import "fmt"

const VERSION = "1.5.1"

func handleVersionCommand(opts *CommandLineOptions, args []string) error {
	fmt.Println(VERSION)
	return nil
}
