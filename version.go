package main

import "fmt"

const VERSION = "1.3.0"

func handleVersionCommand(opts *CommandLineOptions, args []string) error {
	fmt.Println(VERSION)
	return nil
}
