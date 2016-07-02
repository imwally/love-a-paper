package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/imwally/love-a-paper/mdlinks"
)

type Readme struct {
	Path    string
	Content string
}

func IsPDF(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".pdf")
}

func SkipDir(name string) bool {
	prefixes := []string{
		".",
		"_",
	}

	for _, skip := range prefixes {
		if strings.HasPrefix(name, skip) {
			return true
		}
	}

	return false
}

func RandomInt(max int) (int64, error) {
	random, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	return random.Int64(), nil
}

func RandomLink(links []mdlinks.Link) (*mdlinks.Link, error) {
	randInt, err := RandomInt(len(links))
	if err != nil {
		return nil, err
	}
	randLink := links[randInt]

	return &randLink, nil
}

func ScrubScrollNames(links []mdlinks.Link) *[]mdlinks.Link {
	temp := make([]mdlinks.Link, len(links))
	for i, link := range links {
		if link.Name == ":scroll:" {
			link.Name = links[i+1].Name
		}
		temp[i] = link
	}

	return &temp
}

func RandomGithubReadme(owner, repo, dir string) (*Readme, error) {
	if SkipDir(dir) {
		return RandomGithubReadme(owner, repo, "/")
	}

	client := github.NewClient(nil)
	fc, dc, resp, err := client.Repositories.GetContents(owner, repo, dir, nil)
	if err != nil {
		if resp.Remaining < 1 {
			return nil, err
		} else {
			return RandomGithubReadme(owner, repo, "/")
		}
	}

	if fc == nil {
		randInt, err := RandomInt(len(dc))
		if err != nil {
			return nil, err
		}
		randDir := dc[randInt]
		randDirName := randDir.Name

		readmePath := strings.Join([]string{*randDirName, "README.md"}, "/")
		return RandomGithubReadme(owner, repo, readmePath)
	}

	htmlurl := fc.HTMLURL
	path := strings.Replace(*htmlurl, "README.md", "", -1)

	content, err := fc.GetContent()
	if err != nil {
		return nil, err
	}

	return &Readme{path, content}, nil
}

func FindPaper(owner, repo, path string) (*mdlinks.Link, error) {
	readme, err := RandomGithubReadme(owner, repo, path)
	if err != nil {
		return nil, err
	}

	linksUnscrubbed := mdlinks.Links([]byte(readme.Content))
	links := ScrubScrollNames(linksUnscrubbed)

	link, err := RandomLink(*links)
	if err != nil {
		return nil, err
	}

	if !IsPDF(link.Location) {
		return FindPaper(owner, repo, path)
	}

	paperURL, err := url.Parse(link.Location)
	if err != nil {
		return nil, err
	}

	if !paperURL.IsAbs() {
		absURL := strings.Join([]string{readme.Path, link.Location}, "")
		link.Location = absURL
	}

	return link, nil
}

func main() {
	paper, err := FindPaper("papers-we-love", "papers-we-love", "/")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(paper.Name)
	fmt.Println(paper.Location)
}
