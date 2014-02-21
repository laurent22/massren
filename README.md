## Massren

Massren is a command line tool that can be used to rename multiple files using your own text editor. Multiple-rename tools are usually difficult to use from the command line since any regular expression needs to be escaped, and each tool uses its own syntax and flavor of regex. The advantage of massren is that you are using the text editor you use every day, and so can use all its features.

The tool works by creating a file that contains the filenames of the target directory, and opening this file in the text editor. You can then modify the filenames there directly. Once done, save the text file and the files will be renamed. Lines that are not changed will simply be ignored.

![Massren usage animation](https://raw.github.com/laurent22/massren/animation/animation.gif "Massren usage animation")

## Features

- Rename multiple files using your own text editor. It should work with any text editor, including vim, emacs, Sublime Text and even notepad.

- Undo - any rename operation can be undone.

- Dry run mode - test the results of a rename operation without actually renaming any file.

- Cross-platform - Windows, OSX and Linux are supported.

- Safety checks - since this is a multiple rename tool, various checks are in place to ensure that nothing gets accidentally renamed. For example, the program will check that the files are not being changed by something else while the list of filenames is being edited. If the number of files before and after saving the file is different, the operation will also be cancelled.

## Installation

The latest executables for each platform are available from the [release page](https://github.com/laurent22/massren/releases). An installation script is also available:

### OSX

	curl -O https://raw.github.com/laurent22/massren/master/install/install.osx.sh
	chmod 755 install.osx.sh
	sudo ./install.osx.sh

### Linux

	curl -O https://raw.github.com/laurent22/massren/master/install/install.linux-amd64.sh
	chmod 755 install.linux-amd64.sh
	sudo ./install.linux-amd64.sh

### Windows

The executable can be downloaded from https://github.com/laurent22/massren/releases

## Usage and examples

	Usage:
	  massren [OPTIONS] [path]

	Application Options:
	  -n, --dry-run  Don't rename anything but show the operation that would have been performed.
	  -v, --verbose  Enable verbose output.
	  -c, --config   Set a configuration value. eg. massren --config <name> [value]
	  -u, --undo     Undo a rename operation. eg. massren --undo [path]
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

## TODO

- Detect default text editor on Windows.
- Disambiguate filenames when processing two or more folders that contain the same filenames.
- Ability to delete files - see https://github.com/laurent22/massren/issues/11

## Building from source

- Go 1.2+ is required

		go get github.com/laurent22/massren
		go build

## License

MIT
