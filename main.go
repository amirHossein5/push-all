package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var Version = "dev"

type DirEntry struct {
	fullPath string
	fs       fs.DirEntry
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Version:", Version)
		fmt.Println("Usage: push-all <path>")
		os.Exit(1)
	}

	pathToLook := os.Args[1]

	isDir, err := isDir(pathToLook)
	if err != nil {
		log.Fatalln("couldn't open directory:", err)
	}
	if !isDir {
		log.Fatalln("provide a path to directory")
	}

	gitPaths, err := getGitDirsInsideOf(pathToLook)
	if err != nil {
		log.Fatalln("failed to read given path:", err)
	}
	if len(gitPaths) == 0 {
		log.Fatalln("no git projects found in", pathToLook)
	}

	for i, entry := range gitPaths {
		if i != 0 {
			fmt.Println()
		}

		commandString := strings.Join([]string{"cd", entry.fullPath, "&&", "git status && git pull --all && git push"}, " ")

		fmt.Println(">>>", commandString)

		err = streamCommand(commandString)
		if err != nil {
			log.Fatalln("faild to run command", commandString, ":", err)
		}
	}
}

func isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return fileInfo.IsDir(), nil
}

func getDirPathsInsideOf(path string) ([]DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var paths []DirEntry

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		paths = append(paths, DirEntry{
			fullPath: filepath.Join(path, entry.Name()),
			fs:       entry,
		})
	}

	return paths, nil
}

func streamCommand(command string) error {
	cmd := exec.Command("/bin/sh", "-c", command)

	stderr, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	cmd.Start()

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	cmd.Wait()

	return nil
}

func isGitDir(path string) (bool, error) {
	isDir, err := isDir(filepath.Join(path, ".git"))
	if err != nil {
		return false, err
	}

	return isDir, nil
}

func getGitDirsInsideOf(path string) ([]DirEntry, error) {
	dirPaths, err := getDirPathsInsideOf(path)
	if err != nil {
		return nil, err
	}
	if len(dirPaths) == 0 {
		return []DirEntry{}, nil
	}

	var gitDirPaths []DirEntry

	for _, entry := range dirPaths {
		isGitDir, err := isGitDir(entry.fullPath)
		if err != nil {
			return nil, err
		}
		if isGitDir {
			gitDirPaths = append(gitDirPaths, entry)
		}
	}

	return gitDirPaths, nil
}
