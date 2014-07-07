## Massren [![Build Status](https://travis-ci.org/laurent22/massren.png)](https://travis-ci.org/laurent22/massren)

Massren is a command line tool that can be used to rename multiple files using your own text editor. Multiple-rename tools are usually difficult to use from the command line since any regular expression needs to be escaped, and each tool uses its own syntax and flavor of regex. The advantage of massren is that you are using the text editor you use every day, and so can use all its features.

The tool works by creating a file that contains the filenames of the target directory, and opening this file in the text editor. You can then modify the filenames there directly. Once done, save the text file and the files will be renamed. Lines that are not changed will simply be ignored.

![Massren usage animation](https://raw.github.com/laurent22/massren/animation/animation.gif "Massren usage animation")

## Features

- Rename multiple files using your own text editor. It should work with any text editor, including vim, emacs, Sublime Text or notepad.

- Undo - any rename operation can be undone.

- Dry run mode - test the results of a rename operation without actually renaming any file.

- Cross-platform - Windows, OSX and Linux are supported.

- Safety checks - since this is a multiple rename tool, various checks are in place to ensure that nothing gets accidentally renamed. For example, the program will check that the files are not being changed by something else while the list of filenames is being edited. If the number of files before and after saving the file is different, the operation will also be cancelled.

## Installation

The latest executables for each platform are available from the [release page](https://github.com/laurent22/massren/releases). An installation script is also available:

### OSX

#### Homebrew

	brew tap laurent22/massren
	brew install massren

#### With the install script

	curl -O https://raw.github.com/laurent22/massren/master/install/install.osx.sh
	sudo bash install.osx.sh
	
If the installation fails, please follow the [instructions below](#building-from-source).

### Linux

	curl -O https://raw.github.com/laurent22/massren/master/install/install.linux-amd64.sh
	sudo bash install.linux-amd64.sh

### Windows

The executable can be downloaded from https://github.com/laurent22/massren/releases

## Usage and examples

	Usage:
	  massren [OPTIONS]

	Application Options:
	  -n, --dry-run  Don't rename anything but show the operation that would have been performed.
	  -v, --verbose  Enable verbose output.
	  -c, --config   Set or list configuration values. For more info: massren --config --help
	  -u, --undo     Undo a rename operation. Currently delete operations cannot be undone (though files can be recovered from the trash in OSX and Windows). eg. massren --undo [path]
	  -V, --version  Displays version information.

	Help Options:
	  -h, --help     Show this help message

	Examples:

	  Process all the files in the current directory:
	  % massren

	  Process all the JPEGs in the specified directory:
	  % massren /path/to/photos/*.jpg

	  Undo the changes done by the previous operation:
	  % massren --undo /path/to/photos/*.jpg

	  Set VIM as the default text editor:
	  % massren --config editor vim

	  List config values:
	  % massren --config

## Configuration

Type `massren --help --config` (or `massren -ch`) to view the possible configuration values and defaults:

	Config commands:

	  Set a value:
	  % massren --config <name> <value>

	  List all the values:
	  % massren --config

	  Delete a value:
	  % massren --config <name>

	Possible key/values:

	  editor:              The editor to use when editing the list of files.
	                       Default: auto-detected.

	  use_trash:           Whether files should be moved to the trash/recycle bin
	                       after deletion. Possible values: 0 or 1. Default: 1.

	  include_directories: Whether to include the directories the file buffer.
	                       Possible values: 0 or 1. Default: 1.

	  include_header:      Whether to show the header in the file buffer. Possible
	                       values: 0 or 1. Default: 1.

	Examples:

	  Set Sublime as the default text editor:
	  % massren --config editor "subl -n -w"

	  Don't move files to trash:
	  % massren --config use_trash 0

## TODO

- Move files to trash in bulk instead of one by one.
- Detect default text editor on Windows.
- Disambiguate filenames when processing two or more folders that contain the same filenames.
- Other [various issues](https://github.com/laurent22/massren/issues).

## Building from source

- Go 1.2+ is required

		go get github.com/laurent22/massren
		go build
		
- If it doesn't build on OSX, try with:

		go get -x -ldflags -linkmode=external github.com/laurent22/massren

More info in [this issue](https://github.com/laurent22/massren/issues/7).

## License

MIT
