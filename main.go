package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// todo: take git url and clone repository and do work.
// todo: print report in order of likelyhood of a hit. no space +1, upper and lower case +1, has numbers +1, etc
// todo: print file, branch, line info. maybe with a -v flag
// todo: remove duplicates from report, change structure of program to add all lines to data struct
func main() {
	// take command line arg for directory path
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments in call. Use target dir fullpath as args.\n", os.Args[0])
		os.Exit(1)
	}

	directory := os.Args[1]
	var files []string

	err := filepath.Walk(directory, visit(&files))
	if err != nil {
		panic(err)
	}

	// open each file and do work
	// for _, file := range files {
	// 	f, err := os.Open(file)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer f.Close()

	// 	scanner := bufio.NewScanner(f)
	// 	scanner.Split(bufio.ScanLines)

	// 	for scanner.Scan() {
	// 		line := scanner.Text()

	// 		// what are my parse rules?
	// 		// pseudo code for my re
	// 		// an equals sign followed by any amount of whitespace than " with any amount of chars, no whitespace "
	// 		re, err := regexp.Compile(`((?:[^(==)])=[\s'"]*[a-zA-Z0-9\s]*["|'])`)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		if re.Match([]byte(line)) {
	// 			fmt.Println(line)
	// 		}
	// 	}
	// }

	repo, err := git.PlainOpen(directory)
	checkIfError(err)

	// find all branches in repo
	refs, err := repo.Branches()
	checkIfError(err)

	w, err := repo.Worktree()
	checkIfError(err)

	// go to head for each branch
	err = refs.ForEach(func(r *plumbing.Reference) error {
		err = w.Checkout(&git.CheckoutOptions{
			Hash: r.Hash(),
		})
		checkIfError(err)

		ref, err := repo.Head()
		checkIfError(err)

		// fmt.Println(r.Strings())

		// ... retrieves the commit history
		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		checkIfError(err)

		// we want to iterate over every commit, then search all files in that state
		// print out branch name - commit id
		err = cIter.ForEach(func(c *object.Commit) error {
			fmt.Printf("%s -- %s\n", c.Hash, r.Name().Short())

			return nil
		})
		checkIfError(err)

		return nil // ForEach needs some return
	})
	checkIfError(err)

}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		// todo: create a file extention blacklist
		if filepath.Ext(path) == ".js" {
			*files = append(*files, path)
		}
		return nil
	}
}

func checkIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
