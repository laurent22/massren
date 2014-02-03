package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
)

const CONFIG_PERM = 0700

var homeDir_ string
var configFolder_ string

func userHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.HomeDir	
}

func configGet(name string) string {
	b, err := ioutil.ReadFile(configFolder() + "/" + name)
	if err == nil {
		return string(b)
	}
	return ""
}

func configSet(name string, value string) {
	ioutil.WriteFile(configFolder() + "/" + name, []byte(value), CONFIG_PERM)
}

func configDel(name string) {
	os.Remove(configFolder() + "/" + name)
}

func configFolder() string {
	if configFolder_ != "" {
		return configFolder_
	}
	
	if homeDir_ == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		homeDir_ = u.HomeDir
	}
	
	output := homeDir_ + "/.config/" + APPNAME
	
	err := os.MkdirAll(output, CONFIG_PERM)
	if err != nil {
		panic(err)
	}
	
	configFolder_ = output
	return configFolder_
}

func historyFile() string {
	return configFolder() + "/history"
}

func saveHistory(source string, dest string) error {
	f, err := os.OpenFile(historyFile(), os.O_APPEND | os.O_CREATE | os.O_WRONLY, CONFIG_PERM)
	if err != nil {
		return err
	}
	defer f.Close()
	
	_, err = f.WriteString("s " + source + "\n") 
	if err != nil {
		return err
	}
	
	_, err = f.WriteString("d " + dest + "\n") 
	if err != nil {
		return err
	}
	
	return nil
}

func handleConfigCommand(opts *CommandLineOptions, args []string) error {
	if len(args) == 0 {
		return errors.New("no argument specified")
	}
	
	name := args[0]
	
	if len(args) == 1 {
		configDel(name)
		logInfo("Config has been changed: deleted key \"%s\"", name)
		return nil
	}
	
	value := args[1]
	
	configSet(name, value)
	logInfo("Config has been changed: \"%s\" = \"%s\"", name, value)
	return nil
}