package main

import (
	"crypto/rand"
	"log"
	"math/big"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/imwally/love-a-paper/mdlinks"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
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
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
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
// PWL self hosted papers are indicated by a scroll icon prepended to the link,
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
	log.Printf("INFO: scanning %s\n", dir)
	if HasPrefix(dir, []string{".", "_", "CODE_OF_CONDUCT.md"}) {
		log.Printf("INFO: skipping %s\n", dir)
		return RandomGithubReadme(owner, repo, "/")
	}

	client := github.NewClient(nil)
	fc, dc, resp, err := client.Repositories.GetContents(owner, repo, dir, nil)
	log.Printf("GITHUB: %d of %d API requests remaining, reset at %s.", resp.Remaining, resp.Limit, resp.Reset)
	if err != nil {
		if resp.Remaining < 1 {
			return nil, err
		}
		log.Printf("FAILED: %s", err)
		return RandomGithubReadme(owner, repo, "/")
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

	HTMLURL := fc.HTMLURL
	path := strings.Replace(*HTMLURL, "README.md", "", -1)

	content, err := fc.GetContent()
	if err != nil {
		return nil, err
	}

	return &Readme{path, content}, nil
}

// FindPaper is a recursive function that uses RandomGithubReadme to find a
// suitable README.md file. Currently a suitable README.md file is one that
// contains a link to a PDF.
//
// NOTE: Maybe modify IsPDF() to check for other formats such as postscript
// files and rename function to IsPaper().
func FindPaper(owner, repo, path string) (*mdlinks.Link, string, error) {
	readme, err := RandomGithubReadme(owner, repo, path)
	if err != nil {
		return nil, "", err
	}

	linksUnscrubbed := mdlinks.Links([]byte(readme.Content))
	links := ScrubScrollNames(linksUnscrubbed)

	link, err := RandomLink(*links)
	if err != nil {
		return nil, "", err
	}

	if !IsPDF(link.Location) {
		log.Printf("INFO: %s is not a paper", link.Location)
		return FindPaper(owner, repo, path)
	}

	paperURL, err := url.Parse(link.Location)
	if err != nil {
		return nil, "", err
	}

	if !paperURL.IsAbs() {
		absURL := strings.Join([]string{readme.Path, link.Location}, "")
		link.Location = absURL
	}

	topic := strings.Replace(filepath.Base(readme.Path), "_", "", -1)
	link.Name = strings.Replace(link.Name, "\n", " ", -1)

	return link, topic, nil
}

// TwitterLoadCredentials loads Twitter API tokens from environment variables.
func TwitterLoadCredentials() (client *twittergo.Client, err error) {
	config := &oauth1a.ClientConfig{
		ConsumerKey:    os.Getenv("CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("CONSUMER_SECRET"),
	}
	user := oauth1a.NewAuthorizedConfig(os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	client = twittergo.NewClient(config, user)

	return
}

// TwitterUpdateStatus tweets a new status.
func TwitterUpdateStatus(status string) (*twittergo.Tweet, error) {
	client, err := TwitterLoadCredentials()
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("status", status)
	body := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", "/1.1/statuses/update.json", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	tweet := &twittergo.Tweet{}
	err = resp.Parse(tweet)
	if err != nil {
		return nil, err
	}

	return tweet, nil
}

func main() {
	for {
		paper, topic, err := FindPaper("papers-we-love", "papers-we-love", "/")
		if err != nil {
			log.Printf("ERROR: %s\n", err)
		} else {
			log.Printf("INFO: found paper: %s\n", paper.Location)

			status := strings.Join([]string{paper.Name, paper.Location, "#" + topic}, "\n")
			tweet, err := TwitterUpdateStatus(status)
			if err != nil {
				log.Printf("TWITTER: %s\n", err)
			} else {
				log.Printf("TWITTER: tweet successful: %s", tweet.IdStr())
			}
		}

		mrand.Seed(time.Now().Unix())
		// Random integer between 6 and 10. Int(n) returns a random Int from 0
		// to n exclusive.
		sleepTime := mrand.Intn(5) + 6
		log.Printf("INFO: sleeping for %d hours ...", sleepTime)
		time.Sleep(time.Duration(sleepTime) * time.Hour)
	}
}
