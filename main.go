package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	// "time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

// todo: take git url and clone repository and do work.
// todo: print report in order of likelyhood of a hit. no space +1,
// 		 upper and lower case +1 , has numbers +1, str is longer than X +1, etc
// todo: print file, branch, line info. maybe with a -v flag
// todo: remove duplicates from report, change structure of program to add all lines to data struct
// todo: create tests w/ test repo
// maybe need regex for yml, json, toml, etc

func main() {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: os.Args[1],
	})
	checkIfError(err)

	ref, err := r.Head()
	checkIfError(err)

	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	checkIfError(err)

	fmt.Println(commit)

	// TODO 1 - Order prospects
	// step 3 - evaluate diff commit1 commit2

	// ... retrieve the file tree from the commit
	tree, err := commit.Tree()
	checkIfError(err)

	list := map[string][]map[string]string{}
	iter := 0
	lineRe := regexp.MustCompile(`((?:[^(==)])=[\s'"]*[a-zA-Z0-9\s]*["|'])`)
	excludeExtRe := regexp.MustCompile(`^.*\.(jpg|jpeg|tiff|png|JPG|gif|GIF|svg|doc|DOC|pdf|PDF)$`)
	extractRe := regexp.MustCompile(`("|')([^("|')[\]]*)("|')`)

	tree.Files().ForEach(func(file *object.File) error {
		// check for file types we don't want to inspect
		if excludeExtRe.Match([]byte(file.Name)) {
			return nil
		}
		// fmt.Println(file.Contents())
		// fmt.Println("-----------")

		f, err := file.Contents()
		checkIfError(err)

		scanner := bufio.NewScanner(strings.NewReader(f))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := scanner.Text()

			if lineRe.Match([]byte(line)) {
				temp := string(extractRe.Find([]byte(line)))
				ns := strings.Replace(temp, "'", "", -1)
				ns = strings.Replace(ns, "\"", "", -1)
				_, ok := list[ns]
				if !ok {
					// add commit id or branch
					list[ns] = []map[string]string{{"line": line}, {"fileName": file.Name}}
					iter += 1
				}
			}
		}

		for k, _ := range list {
			fmt.Println("k:", k)
		}

		return nil
	})

	fmt.Printf("TOTAL STRINGS = %s\n\n", iter)

	// ... retrieve the tree from the commit
	fmt.Printf("commit parent hashes: %s\n", commit.ParentHashes)

	// since := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	// until := time.Date(2019, 7, 30, 0, 0, 0, 0, time.UTC)
	// commitIter, err := repo.CommitObjects()
	// commitIter, err := r.Log(&git.LogOptions{Order: LogOrderCommitterTime, All: true})
	// commitIter, err := r.Log(&git.LogOptions{From: ref.Hash()})

	// var commits []*Commit
	// object.NewCommitIterCTime(commit, nil, nil).ForEach(func(c *object.Commit) error {
	// 	// commits = append(commits, c)
	// 	fmt.Println(c)
	// 	return nil
	// })
	// checkIfError(err)
}

type hit struct {
	prospect string
	line     string
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		re, err := regexp.Compile(`.*\.(?:jpg|jpeg|tiff|psd|eps|ai|raw|gif|png|bmp|zip)$`)
		if err != nil {
			log.Fatal(err)
		}

		if re.Match([]byte(filepath.Ext(path))) {
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

// Info should be used to describe the example commands that are about to run.
func info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}
