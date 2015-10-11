package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/laurent22/go-sqlkv"
	_ "github.com/mattn/go-sqlite3"
)

const PROFILE_PERM = 0700

var homeDir_ string
var profileFolder_ string
var config_ *sqlkv.SqlKv
var profileDb_ *sql.DB

func userHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.HomeDir
}

func profileOpen() error {
	if profileDb_ != nil {
		return nil
	}

	var err error
	profileDb_, err = sql.Open("sqlite3", profileFile())
	if err != nil {
		return errors.New(fmt.Sprintf("Profile file could not be opened: %s: %s", err, profileFile()))
	}

	_, err = profileDb_.Exec("CREATE TABLE IF NOT EXISTS history (id INTEGER NOT NULL PRIMARY KEY, source TEXT, destination TEXT, timestamp INTEGER)")
	if err != nil {
		return errors.New(fmt.Sprintf("History table could not be created: %s", err))
	}

	profileDb_.Exec("CREATE INDEX id_index ON history (id)")
	profileDb_.Exec("CREATE INDEX destination_index ON history (destination)")
	profileDb_.Exec("CREATE INDEX timestamp_index ON history (timestamp)")

	config_ = sqlkv.New(profileDb_, "config")

	return nil
}

func profileClose() {
	config_ = nil
	if profileDb_ != nil {
		profileDb_.Close()
		profileDb_ = nil
	}
	profileFolder_ = ""
}

// Used only for debugging
func profileDelete() {
	profileFolder := profileFolder_ // profileFolder_ is going to be cleared in profileClose()
	profileClose()
	os.RemoveAll(profileFolder)
}

func profileFolder() string {
	if profileFolder_ != "" {
		return profileFolder_
	}

	if homeDir_ == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		homeDir_ = u.HomeDir
	}

	output := filepath.Join(homeDir_, ".config", APPNAME)

	err := os.MkdirAll(output, PROFILE_PERM)
	if err != nil {
		panic(err)
	}

	profileFolder_ = output
	return profileFolder_
}

func profileFile() string {
	return filepath.Join(profileFolder(), "profile.sqlite")
}

func handleConfigCommand(opts *CommandLineOptions, args []string) error {
	if len(args) == 0 {
		kvs := config_.All()
		for _, kv := range kvs {
			fmt.Printf("%s = \"%s\"\n", kv.Name, kv.Value)
		}
		return nil
	}

	name := args[0]

	if len(args) == 1 {
		config_.Del(name)
		logInfo("Config has been changed: deleted key \"%s\"", name)
		return nil
	}

	value := args[1]

	config_.SetString(name, value)
	logInfo("Config has been changed: %s = \"%s\"", name, value)
	return nil
}
