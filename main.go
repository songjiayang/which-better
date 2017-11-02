package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Version = "0.0.1" // Version is the package version

var (
	r string
	v bool
)

func main() {
	flag.StringVar(&r, "r", "", "the compared repositories split by ,")
	flag.BoolVar(&v, "v", false, "dispaly version")

	flag.Parse()

	if v {
		println("Version is %s", Version)
		return
	}

	if r == "" {
		log.Fatal("Invalid params r")
	}

	splits := strings.Split(r, ",")
	repositories := make([]*Repository, len(splits))
	for i, name := range splits {
		repositories[i] = &Repository{
			Name: name,
		}
	}

	fetch(repositories)
	score(repositories)

	rCollection := RepositoryCollection{
		Repositories: repositories,
	}
	sort.Sort(rCollection)

	for _, r := range rCollection.Repositories {
		r.print()
	}
}

func fetch(repositories []*Repository) {
	var wg sync.WaitGroup

	for _, item := range repositories {
		wg.Add(1)
		go func(r *Repository) {
			defer wg.Done()
			r.fetch()
			r.fetchContributors()
		}(item)
	}

	wg.Wait()
}

func score(repositories []*Repository) {
	size := len(repositories)
	sortedContributors := make([]int, size)
	sortedStars := make([]int, size)
	sortedPushAt := make([]int, size)

	for i, r := range repositories {
		sortedContributors[i] = r.ContributorCount
		sortedStars[i] = r.StarCount
		sortedPushAt[i] = int(r.PushAt.Unix())
	}

	sort.Ints(sortedContributors)
	sort.Ints(sortedStars)
	sort.Ints(sortedPushAt)

	for _, r := range repositories {
		r.Score += (size - sort.SearchInts(sortedContributors, r.ContributorCount))
		r.Score += (size - sort.SearchInts(sortedStars, r.StarCount))
		r.Score += (size - sort.SearchInts(sortedPushAt, int(r.PushAt.Unix())))
	}
}

type RepositoryCollection struct {
	Repositories []*Repository
}

func (this RepositoryCollection) Len() int {
	return len(this.Repositories)
}

func (this RepositoryCollection) Less(i, j int) bool {
	return this.Repositories[i].Score < this.Repositories[j].Score
}

func (this RepositoryCollection) Swap(i, j int) {
	tmp := this.Repositories[i]
	this.Repositories[i] = this.Repositories[j]
	this.Repositories[j] = tmp
}

type Repository struct {
	Name  string `json:"-"`
	Score int

	ContributorCount int
	StarCount        int       `json:"stargazers_count"`
	PushAt           time.Time `json:"pushed_at"`
}

func (this *Repository) print() {
	fmt.Println(fmt.Sprintf("%s: %d", this.Name, this.Score))
}

func (this *Repository) fetch() {
	resp, err := http.Get("https://api.github.com/repos/" + this.Name)

	if err != nil {
		println("fetch(%s): %v", this.Name, err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		println("fetch(%s): %v", this.Name, err)
		return
	}

	json.Unmarshal(body, this)
}

type Contributor struct {
	Id int `id`
}

func (this *Repository) fetchContributors() {
	page := fmt.Sprintf("https://api.github.com/repos/%s/contributors", this.Name)

	nPage, count := this.fetchContributorsWithPage(page, true)
	this.ContributorCount += count

	if nPage != "" {
		_, count = this.fetchContributorsWithPage(nPage, false)
		this.ContributorCount += count
	}
}

var (
	linkRe = regexp.MustCompile(`https.*?page=\d+`)
	pageRe = regexp.MustCompile(`\d+`)
)

func (this *Repository) fetchContributorsWithPage(page string, checkNextPage bool) (nLink string, count int) {
	resp, err := http.Get(page)
	if err != nil {
		println("fetchContributors(%s): %v", this.Name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/200 != 1 {
		println("fetchContributors(%s): %d", this.Name, resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("fetchContributors(%s): %v", this.Name, err)
		return
	}

	var contributors []*Contributor
	err = json.Unmarshal(body, &contributors)
	if err != nil {
		println("fetchContributors(%s): %v", this.Name, err)
		return
	}

	count = len(contributors)

	if checkNextPage && resp.Header.Get("Link") != "" {
		links := linkRe.FindAllString(resp.Header.Get("Link"), -1)
		nLink = links[1]
		lPage := pageRe.FindAllString(nLink, 2)[1]

		if lPage != "" {
			i, err := strconv.Atoi(lPage)
			if err != nil {
				println("fetchContributors(%s): %v", this.Name, err)
				return
			}
			count = count * (i - 1)
		}
	}

	return
}

func println(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}
