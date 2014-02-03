package main

import (
	"fmt"
)

var minLogLevel_ int

func log(level int, s string, a ...interface{}) {
	if level < minLogLevel_ {
		return
	}
	fmt.Printf(APPNAME + ": " + s + "\n", a...)
}

func logDebug(s string, a ...interface{}) {
	log(0, s, a...)
}

func logInfo(s string, a ...interface{}) {
	log(1, s, a...)
}

func logError(s string, a ...interface{}) {
	log(3, s, a...)
}