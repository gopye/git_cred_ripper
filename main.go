package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	// "strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	// fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

// todo: print report in order of likelyhood of a hit. no space +1,
// 		 upper and lower case +1 , has numbers +1, str is longer than X +1, etc
// todo: print file, branch, line info. maybe with a -v flag
// todo: create tests w/ test repo
// maybe need regex for yml, json, toml, etc

// step 4 - evaluate diff commit1 commit2

func main() {
	var dir string
	var url string
	flag.StringVar(&dir, "dir", ".", "repo directory")
	flag.StringVar(&url, "url", "unset", "repo url")
	flag.Parse()

	// fmt.Println(url)

	if url != "unset" {
		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: url,
		})
		checkIfError(err)

		ref, err := r.Head()
		checkIfError(err)

		// ... retrieving the commit object
		commit, err := r.CommitObject(ref.Hash())
		checkIfError(err)

		fmt.Println(commit)

		// ... retrieve the file tree from the commit
		tree, err := commit.Tree()
		checkIfError(err)

		tree.Files().ForEach(func(file *object.File) error {
			prospectCount := 0
			// drop file types we don't want to inspect
			if excludeExtRe.Match([]byte(file.Name)) {
				return nil
			}

			f, err := file.Contents()
			checkIfError(err)

			// scanFileTxt(f)
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
						prospectCount += 1
					}
				}
			}
			return nil
		})
		for k, _ := range list {
			fmt.Println("k:", k)
		}

		fmt.Printf("TOTAL STRINGS = %s\n", len(list))
		// ... retrieve the tree from the commit
		fmt.Printf("commit parent hashes: %s\n", commit.ParentHashes)
	} else if dir != "unset" {

		repo, err := git.PlainOpen(dir)
		checkIfError(err)

		ref, err := repo.Head()
		checkIfError(err)

		// ... retrieving the commit object
		commit, err := repo.CommitObject(ref.Hash())
		checkIfError(err)

		// add commit to work struct
		tempWork = append(tempWork, commit.Hash)
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~")
		fmt.Println(tempWork)
		fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~")

		// pas in work struct
		// scans every file in current branch HEAD
		scanCommitTree(commit, dir)
		checkIfError(err)

		scanParentCommits(commit, repo)	
		

		var keys []string // I think we need this so we can sort.
		for k, _ := range list {
			keys = append(keys, k)
			// list := map[string][]map[string]string{}
			// fmt.Println("k:", k)
			// for _, v1 := range v[1] {
			// 	fmt.Println("v:", strings.TrimSpace(v1))
			// }
			// fmt.Println("------------")
		}

		// should be at very end but don't want to see spam
		sort.Sort(ByLen(keys))
		i := 0
		for i < len(keys) {
			fmt.Println(keys[i])
			i++
		}



		fmt.Println(completed)
		fmt.Printf("\nTOTAL STRINGS = %s\n", len(list))



	} else {
		fmt.Println("Error handling flags")
	}

}

type hit struct {
	prospect string
	line     string
}

var tempWork = make([]plumbing.Hash, 1) // this seems fishy
var work = make(map[plumbing.Hash][]plumbing.Hash)
var completed = make(map[plumbing.Hash][]plumbing.Hash)
var list = map[string][]map[string]string{}
var lineRe = regexp.MustCompile(`((?:[^(==)])=[\s'"]*[a-zA-Z0-9\s]*["|'])`)
var excludeExtRe = regexp.MustCompile(`^.*\.(jpg|jpeg|tiff|png|JPG|gif|GIF|svg|doc|DOC|pdf|PDF)$`)
var extractRe = regexp.MustCompile(`("|')([^("|')[\]]*)("|')`)
var	prospectCount = 0

type ByLen []string

func (a ByLen) Len() int {
	return len(a)
}

func (a ByLen) Less(i, j int) bool{
	return len(a[i]) > len(a[j])
}

func (a ByLen) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
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

func scanCommitTree(c *object.Commit, dir string) error {
	// ... retrieve the tree from the commit
	tree, err := c.Tree()
	checkIfError(err)

	tree.Files().ForEach(func(file *object.File) error {
		// check for file extentions we don't want to inspect
		if excludeExtRe.Match([]byte(file.Name)) {
			return nil
		}

		// Im going to need to have a switch from url v dir param
		fullpath := dir + "/" + file.Name
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
					list[ns] = []map[string]string{{"line": line}, {"commitHash": c.Hash.String()}}
					prospectCount += 1
				}
			}
		}
		return nil
	})
	return nil
}


func scanParentCommits(c *object.Commit, r *git.Repository) {
	fmt.Println(c.ParentHashes)
	// fmt.Println
	if len(c.ParentHashes) == 0 {
		return
	}
	// verify work has not already been done

	// fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	// fmt.Println(completed)
	// fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~")



	for _, parentHash := range c.ParentHashes {
		if completed[c.Hash] != nil {
			fmt.Println("\n\nFOUND \n\n")
			fmt.Println(c.Hash)
			fmt.Printf("%T\n", c.Hash)
			fmt.Println(completed[c.Hash])
			fmt.Printf("%T\n", completed[c.Hash])
			fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			return
		}

		parentCommit, err := r.CommitObject(parentHash)
		checkIfError(err)

		fmt.Println(parentCommit)
		// work[c.Hash] = c.ParentHashes

		// for _, parentHash := range work[c.Hash] {
		// 	parentCommit, err := r.CommitObject(parentHash)
		// 	checkIfError(err)

		parentTree, err := parentCommit.Tree()
		checkIfError(err)
	
		currentTree, err := c.Tree()
		checkIfError(err)

		// Find diff between current and parent trees
		changes, diffErr := object.DiffTree(currentTree, parentTree)
		checkIfError(diffErr)

		if len(changes) == 0 {
			completed[c.Hash] = append(completed[c.Hash], parentHash)
			continue 
		}

		patch, err := changes.Patch()
		checkIfError(err)

		// iterate over these file patches
		for _, filePatch := range patch.FilePatches() {
			for _, v := range filePatch.Chunks() {
				scanner := bufio.NewScanner(strings.NewReader(v.Content()))
				scanner.Split(bufio.ScanLines)

				for scanner.Scan() {
					line := scanner.Text()

					if lineRe.Match([]byte(line)) {
						temp := string(extractRe.Find([]byte(line)))
						ns := strings.Replace(temp, "'", "", -1)
						ns = strings.Replace(ns, "\"", "", -1)
						_, ok := list[ns]
						if !ok {
							list[ns] = []map[string]string{{"line": line}, {"commitHash": parentCommit.Hash.String()}}
							prospectCount += 1
						}
					}
				}
			}
		}
		// add commit -> parentCommit hashes to work done struct
		completed[c.Hash] = append(completed[c.Hash], parentHash)
		scanParentCommits(parentCommit, r)
	}
	// can I recursively call this function?
}
