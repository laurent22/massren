package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setup(t *testing.T) {
	minLogLevel_ = 10

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	homeDir_ = filepath.Join(pwd, "homedirtest")
	err = os.MkdirAll(homeDir_, 0700)
	if err != nil {
		t.Fatal(err)
	}

	deleteTempFiles()
	profileOpen()
	clearHistory()
}

func teardown(t *testing.T) {
	profileDelete()
}

func touch(filePath string) {
	ioutil.WriteFile(filePath, []byte("testing"), 0700)
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func createRandomTempFiles() []string {
	var output []string
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tempFolder(), fmt.Sprintf("testfile%d", i))
		ioutil.WriteFile(filePath, []byte("testing"), 0700)
		output = append(output, filePath)
	}
	return output
}

func fileGetContent(path string) string {
	o, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(o)
}

func filePutContent(path string, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0700)
	if err != nil {
		panic(err)
	}
}

func Test_fileActions(t *testing.T) {
	var err error

	type TestCase struct {
		paths   []string
		content string
		result  []*FileAction
	}

	var testCases []TestCase

	testCases = append(testCases, TestCase{
		paths: []string{
			"abcd",
			"efgh",
			"ijkl",
		},
		content: `
// some header
// some header
// some header

abcd
newname
// should skip this
ijkl
// ignore
`,
		result: []*FileAction{
			&FileAction{
				kind:    KIND_RENAME,
				oldPath: "efgh",
				newPath: "newname",
			},
		},
	})

	testCases = append(testCases, TestCase{
		paths: []string{
			"abcd",
			"efgh",
			"ijkl",
		},
		content: `
// some header
// some header
// some header

//abcd

efgh
ijklmnop
`,
		result: []*FileAction{
			&FileAction{
				kind:    KIND_DELETE,
				oldPath: "abcd",
				newPath: "",
			},
			&FileAction{
				kind:    KIND_RENAME,
				oldPath: "ijkl",
				newPath: "ijklmnop",
			},
		},
	})

	testCases = append(testCases, TestCase{
		paths: []string{
			"abcd",
			"efgh",
			"ijkl",
		},
		content: `
// some header
// some header
abcd
efgh
ijkl
`,
		result: []*FileAction{},
	})

	testCases = append(testCases, TestCase{
		paths: []string{
			" abcd",
			"\t efgh\t\t ",
		},
		content: `
 abcd
	 efgh		 
`,
		result: []*FileAction{},
	})

	testCases = append(testCases, TestCase{
		paths: []string{
			"abcd",
			" efgh",
			" ijkl\t   ",
		},
		content: `
//  abcd
//efgh
// 	 ijkl
`,
		result: []*FileAction{
			&FileAction{
				kind:    KIND_DELETE,
				oldPath: "abcd",
				newPath: "",
			},
			&FileAction{
				kind:    KIND_DELETE,
				oldPath: " efgh",
				newPath: "",
			},
			&FileAction{
				kind:    KIND_DELETE,
				oldPath: " ijkl\t   ",
				newPath: "",
			},
		},
	})

	testCases = append(testCases, TestCase{
		paths: []string{
			" abcd",
			"\t efgh\t\t ",
		},
		content: `
 abcd
	 efgh		 
`,
		result: []*FileAction{},
	})

	testCases = append(testCases, TestCase{
		paths: []string{
			"123",
			"456",
		},
		content: `
1/23
4/56
`,
		result: []*FileAction{
			&FileAction{
				kind:    KIND_RENAME,
				oldPath: "123",
				newPath: "1/23",
			},
			&FileAction{
				kind:    KIND_RENAME,
				oldPath: "456",
				newPath: "4/56",
			},
		},
	})

	// Force \n as newline to simplify testing
	// across platforms.
	newline_ = "\n"

	for _, testCase := range testCases {
		// Note: Run tests with -v in case of error

		r, _ := fileActions(testCase.paths, testCase.content)
		if len(testCase.result) != len(r) {
			t.Errorf("Expected %d, got %d", len(testCase.result), len(r))
			t.Log(testCase.result)
			t.Log(r)
			continue
		}
		for i, r1 := range r {
			r2 := testCase.result[i]
			if r1.kind != r2.kind {
				t.Errorf("Expected kind %d, got %d", r2.kind, r1.kind)
			}
			if r1.oldPath != r2.oldPath {
				t.Errorf("Expected path %s, got %s", r2.oldPath, r1.oldPath)
			}
			if r1.newPath != r2.newPath {
				t.Errorf("Expected path %s, got %s", r2.newPath, r1.newPath)
			}
		}
	}

	_, err = fileActions([]string{"abcd", "efgh"}, "")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	_, err = fileActions([]string{"abcd", "efgh"}, "abcd")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func Test_parseEditorCommand(t *testing.T) {
	type TestCase struct {
		editorCmd  string
		executable string
		args       []string
		hasError   bool
	}

	var testCases []TestCase

	testCases = append(testCases, TestCase{
		editorCmd:  "subl",
		executable: "subl",
		args:       []string{},
		hasError:   false,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "/usr/bin/vim -f",
		executable: "/usr/bin/vim",
		args:       []string{"-f"},
		hasError:   false,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "\"F:\\Sublime Text 3\\sublime_text.exe\" /n /w",
		executable: "F:\\Sublime Text 3\\sublime_text.exe",
		args:       []string{"/n", "/w"},
		hasError:   false,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "subl -w --command \"something with spaces\"",
		executable: "subl",
		args:       []string{"-w", "--command", "something with spaces"},
		hasError:   false,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "notepad.exe /PT",
		executable: "notepad.exe",
		args:       []string{"/PT"},
		hasError:   false,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "vim -e \"unclosed quote",
		executable: "",
		args:       []string{},
		hasError:   true,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "subl -e 'unclosed single-quote",
		executable: "",
		args:       []string{},
		hasError:   true,
	})

	testCases = append(testCases, TestCase{
		editorCmd:  "",
		executable: "",
		args:       []string{},
		hasError:   true,
	})

	for _, testCase := range testCases {
		executable, args, err := parseEditorCommand(testCase.editorCmd)
		if (err != nil && !testCase.hasError) || (err == nil && testCase.hasError) {
			t.Errorf("Error status did not match: %t: %s", testCase.hasError, err)
		}
		if executable != testCase.executable {
			t.Errorf("Expected '%s', got '%s'", testCase.executable, executable)
		}
		if len(args) != len(testCase.args) {
			t.Errorf("Expected and result args don't have the same length: [%s], [%s]", strings.Join(testCase.args, ", "), strings.Join(args, ", "))
		}
		for i, arg := range testCase.args {
			if arg != args[i] {
				t.Errorf("Expected and result args differ: [%s], [%s]", strings.Join(testCase.args, ", "), strings.Join(args, ", "))
			}
		}
	}
}

func Test_processFileActions(t *testing.T) {
	setup(t)
	defer teardown(t)

	touch(filepath.Join(tempFolder(), "one"))
	touch(filepath.Join(tempFolder(), "two"))
	touch(filepath.Join(tempFolder(), "three"))

	fileActions := []*FileAction{}

	fileAction := NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "one")
	fileAction.newPath = "one123"
	fileActions = append(fileActions, fileAction)

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "two")
	fileAction.newPath = "two456"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, false)

	if !fileExists(filepath.Join(tempFolder(), "one123")) {
		t.Error("File not found")
	}

	if !fileExists(filepath.Join(tempFolder(), "two456")) {
		t.Error("File not found")
	}

	if !fileExists(filepath.Join(tempFolder(), "three")) {
		t.Error("File not found")
	}

	fileActions = []*FileAction{}

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "two456")
	fileAction.kind = KIND_DELETE
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, true)

	if !fileExists(filepath.Join(tempFolder(), "two456")) {
		t.Error("File should not have been deleted")
	}

	processFileActions(fileActions, false)

	if fileExists(filepath.Join(tempFolder(), "two456")) {
		t.Error("File should have been deleted")
	}

	fileActions = []*FileAction{}

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "three")
	fileAction.newPath = "nochange"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, true)

	if !fileExists(filepath.Join(tempFolder(), "three")) {
		t.Error("File was renamed in dry-run mode")
	}

	fileActions = []*FileAction{}

	touch(filepath.Join(tempFolder(), "abcd"))

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "abcd")
	fileAction.newPath = "ab/cd"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, false)

	if !fileExists(filepath.Join(tempFolder(), "ab", "cd")) {
		t.Error("File was not renamed or not moved to subfolder")
	}
}

func Test_processFileActions_noDestinationOverwrite(t *testing.T) {
	setup(t)
	defer teardown(t)

	// Case where a sequence of files such as 0.jpg, 1.jpg is being
	// renamed to 1.jpg, 2.jpg

	p0 := filepath.Join(tempFolder(), "0")
	p1 := filepath.Join(tempFolder(), "1")
	p2 := filepath.Join(tempFolder(), "2")

	filePutContent(p0, "0")
	filePutContent(p1, "1")

	fileActions := []*FileAction{}

	fileAction := NewFileAction()
	fileAction.oldPath = p0
	fileAction.newPath = "1"
	fileActions = append(fileActions, fileAction)

	fileAction = NewFileAction()
	fileAction.oldPath = p1
	fileAction.newPath = "2"
	fileActions = append(fileActions, fileAction)

	err := processFileActions(fileActions, false)

	if err != nil {
		t.Errorf("Expected no error, but got an error.")
	}

	if fileExists(p0) {
		t.Error("File 0 should have been renamed")
	}

	if !fileExists(p1) {
		t.Error("File 1 should exist")
	}

	if !fileExists(p2) {
		t.Error("File 2 should exist")
	}

	if fileGetContent(p1) != "0" {
		t.Error("File 1 has wrong content")
	}

	if fileGetContent(p2) != "1" {
		t.Error("File 2 has wrong content")
	}
}

func Test_processFileActions_noDestinationOverwrite2(t *testing.T) {
	setup(t)
	defer teardown(t)

	// Case where a file is renamed to the name of
	// another existing file.

	touch(filepath.Join(tempFolder(), "0"))
	touch(filepath.Join(tempFolder(), "1"))

	originalFilePaths := []string{
		filepath.Join(tempFolder(), "0"),
	}

	changes := `
1
`
	_, err := fileActions(originalFilePaths, changes)

	if err == nil {
		t.Error("Expected an error, but got nil.")
	}
}

func Test_processFileActions_swapFilenames(t *testing.T) {
	setup(t)
	defer teardown(t)

	p0 := filepath.Join(tempFolder(), "0")
	p1 := filepath.Join(tempFolder(), "1")

	filePutContent(p0, "0")
	filePutContent(p1, "1")

	originalFilePaths := []string{
		p0,
		p1,
	}

	changes := `
1
0
`
	actions, _ := fileActions(originalFilePaths, changes)
	err := processFileActions(actions, false)

	if err != nil {
		t.Error("Expected no error, but got one.")
	}

	if fileGetContent(p0) != "1" {
		t.Error("File 1 has not been renamed correctly")
	}

	if fileGetContent(p1) != "0" {
		t.Error("File 0 has not been renamed correctly")
	}
}

func Test_processFileActions_renameToSameName(t *testing.T) {
	setup(t)
	defer teardown(t)

	touch(filepath.Join(tempFolder(), "0"))
	touch(filepath.Join(tempFolder(), "1"))

	originalFilePaths := []string{
		filepath.Join(tempFolder(), "0"),
		filepath.Join(tempFolder(), "1"),
	}

	changes := `
9
9
`
	_, err := fileActions(originalFilePaths, changes)

	if err == nil {
		t.Error("Expected an error, but got nil.")
	}
}

func Test_processFileActions_dontDeleteAfterRename(t *testing.T) {
	setup(t)
	defer teardown(t)

	p0 := filepath.Join(tempFolder(), "0")
	p1 := filepath.Join(tempFolder(), "1")
	filePutContent(p0, "0")
	filePutContent(p1, "1")

	originalFilePaths := []string{
		p0,
		p1,
	}

	changes := `
1
//1
`
	actions, _ := fileActions(originalFilePaths, changes)
	err := processFileActions(actions, false)

	if err != nil {
		t.Error("Expected no error, but got one.")
	}

	if fileExists(p0) {
		t.Error("File 0 should have been renamed but still exists.")
	}

	if !fileExists(p1) {
		t.Error("File 1 should exist.")
	}

	if fileGetContent(p1) != "0" {
		t.Errorf("File 0 has not been renamed correctly - content should be \"%s\", but is \"%s\"", "0", fileGetContent(p1))
	}
}

func Test_stringHash(t *testing.T) {
	if len(stringHash("aaaa")) != 32 {
		t.Error("hash should be 32 characters long")
	}

	if stringHash("abcd") == stringHash("efgh") || stringHash("") == stringHash("ijkl") {
		t.Error("hashes should be different")
	}
}

func Test_watchFile(t *testing.T) {
	setup(t)
	defer teardown(t)

	filePath := tempFolder() + "watchtest"
	ioutil.WriteFile(filePath, []byte("testing"), 0700)
	doneChan := make(chan bool)

	go func(doneChan chan bool) {
		defer func() {
			doneChan <- true
		}()
		err := watchFile(filePath)
		if err != nil {
			t.Error(err)
		}
	}(doneChan)

	time.Sleep(300 * time.Millisecond)
	ioutil.WriteFile(filePath, []byte("testing change"), 0700)

	<-doneChan
}

func fileListsAreEqual(files1 []string, files2 []string) error {
	if len(files1) != len(files2) {
		return errors.New("file count is different")
	}

	for _, f1 := range files1 {
		found := false
		for _, f2 := range files2 {
			if filepath.Base(f1) == filepath.Base(f2) {
				found = true
			}
		}
		if !found {
			return errors.New("file names do not match")
		}
	}

	return nil
}

func Test_filePathsFromArgs(t *testing.T) {
	setup(t)
	defer teardown(t)

	tempFiles := createRandomTempFiles()
	args := []string{
		filepath.Join(tempFolder(), "*"),
	}

	filePaths, err := filePathsFromArgs(args, true)
	if err != nil {
		t.Fatal(err)
	}

	err = fileListsAreEqual(filePaths, tempFiles)
	if err != nil {
		t.Error(err)
	}

	// If no argument is provided, the function should default to "*"
	// in the current dir.

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(tempFolder())
	if err != nil {
		panic(err)
	}

	args = []string{}
	filePaths, err = filePathsFromArgs(args, true)
	if err != nil {
		t.Fatal(err)
	}

	err = fileListsAreEqual(filePaths, tempFiles)
	if err != nil {
		t.Error(err)
	}

	os.Chdir(currentDir)
}

func Test_filePathsFromArgs_noDirectories(t *testing.T) {
	setup(t)
	defer teardown(t)

	f0 := filepath.Join(tempFolder(), "0")
	f1 := filepath.Join(tempFolder(), "1")
	d0 := filepath.Join(tempFolder(), "dir0")
	d1 := filepath.Join(tempFolder(), "dir1")

	touch(f0)
	touch(f1)
	os.Mkdir(d0, 0700)
	os.Mkdir(d1, 0700)

	args := []string{
		filepath.Join(tempFolder(), "*"),
	}

	filePaths, _ := filePathsFromArgs(args, true)
	err := fileListsAreEqual(filePaths, []string{f0, f1, d0, d1})
	if err != nil {
		t.Error(err)
	}

	filePaths, _ = filePathsFromArgs(args, false)
	err = fileListsAreEqual(filePaths, []string{f0, f1})
	if err != nil {
		t.Error(err)
	}
}

func stringListsEqual(s1 []string, s2 []string) bool {
	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}
	return true
}

func Test_filePathsFromString(t *testing.T) {
	newline_ = "\n"

	var data []string
	var expected [][]string

	data = append(data, "// comment\n\nfile1\nfile2\n//comment\n\n\n")
	expected = append(expected, []string{"file1", "file2"})

	data = append(data, "\n// comment\n\n")
	expected = append(expected, []string{})

	data = append(data, "")
	expected = append(expected, []string{})

	data = append(data, "// comment\n\n  file1 \n\tfile2\n\nanother file\t\n//comment\n\n\n")
	expected = append(expected, []string{"  file1 ", "\tfile2", "another file\t"})

	for i, d := range data {
		e := expected[i]
		r := filePathsFromString(d)
		if !stringListsEqual(e, r) {
			t.Error("Expected", e, "got", r)
		}
	}
}

func Test_filePathsFromListFile(t *testing.T) {
	setup(t)
	defer teardown(t)

	ioutil.WriteFile(filepath.Join(tempFolder(), "list.txt"), []byte("one"+newline()+"two"), PROFILE_PERM)
	filePaths, err := filePathsFromListFile(filepath.Join(tempFolder(), "list.txt"))
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if len(filePaths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(filePaths))
	} else {
		if filePaths[0] != "one" || filePaths[1] != "two" {
			t.Error("Incorrect data")
		}
	}

	os.Remove(filepath.Join(tempFolder(), "list.txt"))
	_, err = filePathsFromListFile(filepath.Join(tempFolder(), "list.txt"))
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func Test_stripBom(t *testing.T) {
	data := [][][]byte{
		{{239, 187, 191}, {}},
		{{239, 187, 191, 239, 187, 191}, {239, 187, 191}},
		{{239, 187, 191, 65, 66}, {65, 66}},
		{{239, 191, 65, 66}, {239, 191, 65, 66}},
		{{}, {}},
		{{65, 239, 187, 191}, {65, 239, 187, 191}},
	}

	for _, d := range data {
		if stripBom(string(d[0])) != string(d[1]) {
			t.Errorf("Expected %x, got %x", d[0], d[1])
		}
	}
}

func Test_deleteTempFiles(t *testing.T) {
	setup(t)
	defer teardown(t)

	ioutil.WriteFile(filepath.Join(tempFolder(), "one"), []byte("test1"), PROFILE_PERM)
	ioutil.WriteFile(filepath.Join(tempFolder(), "two"), []byte("test2"), PROFILE_PERM)

	deleteTempFiles()

	tempFiles, _ := filepath.Glob(filepath.Join(tempFolder(), "*"))
	if len(tempFiles) > 0 {
		t.Fail()
	}
}

func Test_newline(t *testing.T) {
	newline_ = ""
	nl := newline()
	if len(nl) < 1 || len(nl) > 2 {
		t.Fail()
	}
}

func Test_guessEditorCommand(t *testing.T) {
	editor, err := guessEditorCommand()
	if err != nil || len(editor) <= 0 {
		t.Fail()
	}
}

func Test_bufferHeader(t *testing.T) {
	setup(t)
	defer teardown(t)

	f0 := filepath.Join(tempFolder(), "0")
	f1 := filepath.Join(tempFolder(), "1")

	touch(f0)
	touch(f1)

	newline_ = "\n"

	content := createListFileContent([]string{f0, f1}, true)
	if strings.Index(content, "//") != 0 {
		t.Fatal("cannot find header")
	}

	content = createListFileContent([]string{f0, f1}, false)
	if content != "0\n1\n" {
		t.Fatal("file content is incorrect: " + content)
	}
}
