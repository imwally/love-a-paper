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

// Readme holds the path to and content of the README file found in a github
// repository directory.
type Readme struct {
	Path    string
	Content string
}

// IsPDF returns true if the path has a .pdf extention. It is case insensitive.
func IsPDF(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".pdf")
}

// HasPrefix returns true if name starts with any string found in the slice
// of prefixes.
func HasPrefix(name string, prefixes []string) bool {
	for _, skip := range prefixes {
		if strings.HasPrefix(name, skip) {
			return true
		}
	}

	return false
}

// RandomInt returns a random int64 between 0 and max.
func RandomInt(max int) (int64, error) {
	random, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	return random.Int64(), nil
}

// RandomLink returns a random mdlinks.Link from a slice.
func RandomLink(links []mdlinks.Link) (*mdlinks.Link, error) {
	randInt, err := RandomInt(len(links))
	if err != nil {
		return nil, err
	}
	randLink := links[randInt]

	return &randLink, nil
}

// ScrubScrollNames is a helper function that replaces any Link containing the
// name ":scroll:" with the name of the Link found directly after it. Because
// PWL self hosted papers are indicated by a scroll icon appended to the link,
// the markdown parser assumes this is the name of the link. In most cases the
// actual name is found immediately after the scroll icon.
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

// RandomGithubReadme returns a README file from a randomly chosen directory
// within a Github repository. It will recursively try to find a README file
// until either one is found or the Github API rate limit has been hit.
func RandomGithubReadme(owner, repo, dir string) (*Readme, error) {
	// Start again from the root directory if the current directory has any of
	// the following prefixes.
	if HasPrefix(dir, []string{".", "_"}) {
		return RandomGithubReadme(owner, repo, "/")
	}

	client := github.NewClient(nil)
	fc, dc, resp, err := client.Repositories.GetContents(owner, repo, dir, nil)
	if err != nil {
		// If the Github API rate limit has been hit return with the error.
		if resp.Remaining < 1 {
			return nil, err
			// Otherwise a 404 (not a README.md file) or other HTTP error occured.
			// Start again from the root directory.
		} else {
			return RandomGithubReadme(owner, repo, "/")
		}
	}

	// No README.md file has been found yet so check the current directoy for
	// README.md.
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

	// Remember the path where the README.md file was found.
	htmlurl := fc.HTMLURL
	path := strings.Replace(*htmlurl, "README.md", "", -1)

	// Extract content from the README.md file.
	content, err := fc.GetContent()
	if err != nil {
		return nil, err
	}

	return &Readme{path, content}, nil
}

// FindPaper is a recursive function that uses RandomGithubReadme to find a
// suitable README.md file containing a link to a paper. Currently a suitable
// README.md file is one that contains a link to a PDF.
//
// NOTE: Maybe modify IsPDF() to check for other formats such as postscript
// files and rename function to IsPaper().
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
		return
	}

	fmt.Println(paper.Name)
	fmt.Println(paper.Location)
}
