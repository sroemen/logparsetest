package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var (
	accountInfo = map[string]actStats{}
)

type actStats struct {
	AccountID string
	Instances []int64
	PageHits  int64
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

			// skip directories
			if info.IsDir() {
				return nil
			}

			// parse the file(s)
			return parseFile(path)
		})
	if err != nil {
		log.Println(err)
	}

	// Print out the results.
	printStats()
}

func parseFile(fpath string) error {

	file, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var datepath = regexp.MustCompile(`[0-9]+/[a-zA-Z]+/[0-9]+:[0-9]+:[0-9]+:[0-9]+ [+-][0-9]+`)
	var urlpath = regexp.MustCompile(`/[a-zA-Z0-9]+`)

	// read each line and process the data
	for scanner.Scan() {
		// the regex finds the / paths.  slice 1 = year 2 = month
		//  Hence we start > 2
		urlsplit := urlpath.FindAll(scanner.Bytes(), -1)[2:]
		date1 := datepath.FindAll(scanner.Bytes(), -1)

		if len(urlsplit) > 2 {
			id := string(urlsplit[2])[1:]
			if len(id) > 6 {

				const (
					logLayout = "02/Jan/2006:15:04:05 -0700"
				)

				thetime, e := time.Parse(logLayout, fmt.Sprintf("%s", date1[0]))
				if e != nil {
					fmt.Println(e)
				}

				// create the Instances info
				ai, found := accountInfo[id]
				if found {
					ai.Instances = append(ai.Instances, thetime.Unix())
				}

				// Add to the map
				accountInfo[id] = actStats{AccountID: id, PageHits: accountInfo[id].PageHits + 1, Instances: ai.Instances}
			}
		} else {
			// enable to show what didn't match the urlsplit regex (ELBs)
			//fmt.Printf("failed parsing: %s\n", scanner.Text())
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
	fmt.Printf("Total unique users: %v\n", len(accountInfo))
	fmt.Println("Top users:")
	fmt.Printf("id\t\t# pages\t# sess\tlongest\tshortest\n")

	d1 := []*kv{}
	for _, d := range accountInfo {
		d1 = append(d1, &kv{d.AccountID, d.PageHits})
	}
	// now sort d1, and get the top 5 accounts
	sort.Sort(sort.Reverse(byHits(d1)))

	// list out just the first 5 entries
	for _, as := range d1[:5] {
		sess, long, short := calculateSessions(accountInfo[as.ID].Instances)
		fmt.Printf("%s\t%v\t%v\t%v\t%v\n", accountInfo[as.ID].AccountID, accountInfo[as.ID].PageHits, sess, long, short)
	}
}

func calculateSessions(i []int64) (int64, int64, int64) {
	sess := int64(1)
	long := int64(0)
	short := int64(60) //set min to 1 minute

	start := int64(0)
	sessionTime := int64(0)

	for num, v := range i {
		// don't start sessionTime from the first entry (nothing to subtract from, causing large deltas)
		if num != 0 {
			// determine the delta
			sessionTime = sessionTime + (v - start)
		}

		if sessionTime > long {
			long = sessionTime
		} else {
			if sessionTime < short && sessionTime > 0 {
				short = sessionTime
			}
		}

		if (v-start) > 600 && num != 0 {
			// increment session by one
			sess = sess + 1
			// reset sessionTime to zero, since it's a new session.
			sessionTime = 0
			long = 0
			short = 60
		}

		// set the start to the current timestamp
		start = v
	}

	// sessions don't last 0 seconds (i'd hope) so set short to long if only one session
	if sess == 1 {
		// if longest time is less than a minute, make it a minute for stats
		if long < 60 {
			long = 60
		}
		short = long
	}

	// divide long, short by 60, to return minutes
	return sess, long / 60, short / 60
}
