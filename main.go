package main

import (
	"bufio"
	"fmt"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sort"

	// "github.com/go-git/go-git"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
			iter := 0
			// drop file types we don't want to inspect
			if excludeExtRe.Match([]byte(file.Name)) {
				return nil
			}

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

		scanCommitTree(commit, dir)
		checkIfError(err)


		var keys []string
		for k, _ := range list {
			keys = append(keys, k)
			// list := map[string][]map[string]string{}
			// fmt.Println("k:", k)
			// for _,v1 := range v[0] {
			// 	fmt.Println("v:", strings.TrimSpace(v1))
			// }
			// fmt.Println("------------")
		}

		fmt.Printf("TOTAL STRINGS = %s\n", len(list))
		
		sort.Sort(ByLen(keys))

		i := 0
		for i < len(keys) {
			fmt.Println(keys[i])
			i++
		}

		var work = make(map[plumbing.Hash][]plumbing.Hash)
		work[commit.Hash] = commit.ParentHashes
		fmt.Println(commit.Hash)
		fmt.Println(commit.ParentHashes)

		fmt.Println("----------------" )
		fmt.Println(work)

		work[commit.Hash] = append(work[commit.Hash], commit.Hash)
		fmt.Println("----------------" )
		fmt.Println(work)
	} else {
		fmt.Println("Error handling flags")
	}

}

type hit struct {
	prospect string
	line     string
}

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

var list = map[string][]map[string]string{}
var lineRe = regexp.MustCompile(`((?:[^(==)])=[\s'"]*[a-zA-Z0-9\s]*["|'])`)
var excludeExtRe = regexp.MustCompile(`^.*\.(jpg|jpeg|tiff|png|JPG|gif|GIF|svg|doc|DOC|pdf|PDF)$`)
var extractRe = regexp.MustCompile(`("|')([^("|')[\]]*)("|')`)

func scanCommitTree(c *object.Commit, dir string) error {
	iter := 0
	// ... retrieve the tree from the commit
	tree, err := c.Tree()
	if err != nil {
		return err
	}

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
					// add commit id or branch
					list[ns] = []map[string]string{{"line": line}, {"fileName": file.Name}}
					iter += 1
				}
			}
		}

		return nil
	})

	return nil
}