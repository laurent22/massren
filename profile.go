package main

import (
	"database/sql"
	"errors"
	"os"
	"os/user"
	
	"github.com/laurent22/go-sqlkv"
    _ "github.com/mattn/go-sqlite3" 
)

const CONFIG_PERM = 0700

var homeDir_ string
var configFolder_ string
var config_ *sqlkv.SqlKv
var profileDb_ *sql.DB

func userHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.HomeDir
}

func profileOpen() {
	if profileDb_ != nil {
		return
	}
	
	logDebug("Opening profile...")
	
	var err error
	profileDb_, err = sql.Open("sqlite3", profileFile())
	if err != nil {
		logError("Profile file could not be opened: %s: %s", err, profileFile())
	}

	_, err = profileDb_.Exec("CREATE TABLE IF NOT EXISTS history (id INTEGER NOT NULL PRIMARY KEY, source TEXT, destination TEXT, timestamp INTEGER)")
	if err != nil {
		logError("History table could not be created: %s", err)
	}
	
	profileDb_.Exec("CREATE INDEX id_index ON history (id)")
	profileDb_.Exec("CREATE INDEX destination_index ON history (destination)")
	profileDb_.Exec("CREATE INDEX timestamp_index ON history (timestamp)")

	config_ = sqlkv.New(profileDb_, "config")
}

func profileClose() {
	logDebug("Closing profile...")

	config_ = nil
	if profileDb_ != nil {
		profileDb_.Close()
		profileDb_ = nil
	}
}

func profileFolder() string {
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

func profileFile() string {
	return profileFolder() + "/profile.sqlite"
}

func handleConfigCommand(opts *CommandLineOptions, args []string) error {
	if len(args) == 0 {
		return errors.New("no argument specified")
	}
	
	name := args[0]
	
	if len(args) == 1 {
		config_.Del(name)
		logInfo("Config has been changed: deleted key \"%s\"", name)
		return nil
	}
	
	value := args[1]
	
	config_.SetString(name, value)
	logInfo("Config has been changed: \"%s\" = \"%s\"", name, value)
	return nil
}