package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var (
	accountInfo = map[string]actStats{}
)

type actStats struct {
	AccountID       string
	Sesssions       []string
	PageHits        int64
	LongestSession  int64
	ShortestSession int64
}

func main() {
	// Exit out if path to the logs was not passed.
	if len(os.Args) < 2 {
		log.Fatal("Application takes one argument, the path to the logs.")
	}

	// take the first argument passed to the app as the log dir
	logdir := os.Args[1]

	accountInfo = make(map[string]actStats, 0)

	// walk the directory, and process each file
	err := filepath.Walk(logdir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			//skip directories
			if info.IsDir() {
				return nil
			}

			fmt.Println(path, info.Size())

			return parseFile(path)
		})
	if err != nil {
		log.Println(err)
	}

	printStats()
}

func parseFile(fpath string) error {

	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var urlpath = regexp.MustCompile(`/[a-zA-Z0-9]+`)

	// read each line and process the data
	for scanner.Scan() {
		// the regex finds the / paths.  slice 1 = year 2 = month
		//  Hence we start > 2
		urlsplit := urlpath.FindAll(scanner.Bytes(), -1)[2:]

		if len(urlsplit) > 2 {
			id := string(urlsplit[2])[1:]
			if len(id) > 6 {
				accountInfo[id] = actStats{AccountID: id, PageHits: accountInfo[id].PageHits + 1}

				fmt.Printf("%s \n", urlsplit)

				fmt.Println(scanner.Text())
			}
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil
}

type kv struct {
	ID   string
	Hits int64
}

// actSort implements sort.Interface for the PageHits field
type byHits []*kv

func (a byHits) Len() int           { return len(a) }
func (a byHits) Less(x, y int) bool { return a[x].Hits < a[y].Hits }
func (a byHits) Swap(x, y int)      { a[x], a[y] = a[y], a[x] }

func printStats() {

	/*
		Total unique users: 27
		Top users:
		id              # pages # sess  longest shortest
		71f28176        75      3       35      1
		41f58122        65      4       60      10
		58122233        44      2       121     3
	*/
	fmt.Printf("\n\nTotal unique users: %v\n", len(accountInfo))
	fmt.Println("Top users:")
	fmt.Printf("id\t\t# pages\t# sess\tlongest\tshortest\n")

	// to be replaced by a loop of the real stats
	//fmt.Printf("71f28176\t75\t3\t35\t1\n")

	d1 := []*kv{}
	for _, d := range accountInfo {
		d1 = append(d1, &kv{d.AccountID, d.PageHits})
	}
	// now sort d1, and get the top 5 accounts
	sort.Sort(sort.Reverse(byHits(d1)))

	for _, as := range d1[:5] {
		fmt.Printf("%s\t%v\t3\t35\t1\n", accountInfo[as.ID].AccountID, accountInfo[as.ID].PageHits)
	}
}
