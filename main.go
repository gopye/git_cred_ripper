package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	// take command line arg for directory path
	// todo: take git url and clone repository and do work
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments in call. Use target dir fullpath as args.\n", os.Args[0])
		os.Exit(1)
	}

	root := os.Args[1]
	var files []string

	err := filepath.Walk(root, visit(&files))
	if err != nil {
		panic(err)
	}

	// open each file and do work
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanLines)

		// remove duplicates, change structure of program to add all lines to data struct
		for scanner.Scan() {
			line := scanner.Text()

			// what are my parse rules?
			// pseudo code for my re
			// an equals sign followed by any amount of whitespace than " with any amount of chars, no whitespace "
			re, err := regexp.Compile(`((?:[^(==)])=[\s'"]*[a-zA-Z0-9\s]*["|'])`)
			if err != nil {
				log.Fatal(err)
			}
			if re.Match([]byte(line)) {
				fmt.Println(line)
			}
		}
	}

	// open git and iterate over it
	// adv go through each file -- add vars to [] remove duplicates
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		// make this more inclusive
		// come up with a regex to exclude photo blobs
		if filepath.Ext(path) == ".js" {
			*files = append(*files, path)
		}
		return nil
	}
}
