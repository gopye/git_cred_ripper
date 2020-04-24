package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	// "github.com/go-git/go-git/v4"
	// "github.com/go-git/go-git/v4/plumbing/object"
)

// todo: take git url and clone repository and do work.
// todo: print report in order of likelyhood of a hit. no space +1,
// 		 upper and lower case +1 , has numbers +1, str is longer than X +1, etc
// todo: print file, branch, line info. maybe with a -v flag
// todo: remove duplicates from report, change structure of program to add all lines to data struct
// todo: create tests w/ test repo
// maybe need regex for yml, json, toml, etc

func main() {
	// get command line arg for directory path
	if len(os.Args) != 2 {
		fmt.Println("Not enough arguments in call. Use target dir fullpath as args.\n", os.Args[0])
		os.Exit(1)
	}

	directory := os.Args[1]

	repo, err := git.PlainOpen(directory)
	checkIfError(err)

	ref, err := repo.Head()
	checkIfError(err)

	// ... retrieving the commit object
	commit, err := repo.CommitObject(ref.Hash())
	checkIfError(err)

	fmt.Println(commit)

	// TODO 1 - Order prospects
	// step 3 - evaluate diff commit1 commit2

	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	checkIfError(err)

	fmt.Println("\n\nIGNORE\n\n")

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

		fullpath := directory + "/" + file.Name
		f, err := os.Open(fullpath)
		if err != nil {
			fmt.Printf("CANNOT OPEN: %s", file.Name)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
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
	// List the tree from HEAD
	info("git ls-tree -r HEAD")

	// ... retrieve the tree from the commit
	fmt.Printf("commit parent hashes: %s\n", commit.ParentHashes)

	// List the history of the repository
	info("git log --oneline")

	// commitIter, err := repo.CommitObjects()
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash(), All: true})
	// commitIter, err := repo.Log(&git.LogOptions{Order: LogOrderCommitterTime, All: true})
	checkIfError(err)
	err = commitIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)

		return nil
	})

	// find all branches in repo
	// refs, err := repo.Branches()
	// checkIfError(err)

	// w, err := repo.Worktree()
	// checkIfError(err)

	// maybe just iterate over every commit w/ func (r *Repository) CommitObjects()
	// iterate over each branch
	// err = refs.ForEach(func(r *plumbing.Reference) error {
	// 	err = w.Checkout(&git.CheckoutOptions{
	// 		Hash: r.Hash(),
	// 	})
	// 	checkIfError(err)

	// 	ref, err := repo.Head()
	// 	checkIfError(err)

	// we want to iterate over every commit, then search all files in that state
	// err = commitIter.ForEach(func(c *object.Commit) error {
	// fmt.Printf("%s -- %s\n", c.Hash, r.Name().Short())
	// }

	// for _, file := range files {
	// 	f, err := os.Open(file)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer f.Close()

	// 	fmt.Println(f)
	// }
	// 		return nil
	// 	})
	// 	checkIfError(err)

	// 	return nil
	// })
	checkIfError(err)
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
