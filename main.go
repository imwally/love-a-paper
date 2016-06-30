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

func IsPdf(name string) bool {
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

func RandomInt(maxInt int) (int64, error) {
	max := big.NewInt(int64(maxInt))
	random, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return random.Int64(), nil
}

func RandomReadme(dir string) (*Readme, error) {
	pwl := "papers-we-love"

	if SkipDir(dir) {
		return RandomReadme("/")
	}

	client := github.NewClient(nil)
	fc, dc, resp, err := client.Repositories.GetContents(pwl, pwl, dir, nil)
	if err != nil {
		return nil, err
	}
	log.Println(resp)

	if fc == nil {
		randInt, err := RandomInt(len(dc))
		if err != nil {
			return nil, err
		}
		randDir := dc[randInt]
		randDirName := randDir.Name

		readmePath := strings.Join([]string{*randDirName, "README.md"}, "/")
		return RandomReadme(readmePath)
	}

	htmlurl := fc.HTMLURL
	path := strings.Replace(*htmlurl, "README.md", "", -1)

	content, err := fc.GetContent()
	if err != nil {
		return nil, err
	}

	return &Readme{path, content}, nil
}

func main() {
	readme, err := RandomReadme("/")
	if err != nil {
		log.Println(err)
	}

	links := mdlinks.Links([]byte(readme.Content))

	for _, link := range links {
		if IsPdf(link.Location) {
			paperURL, err := url.Parse(link.Location)
			if err != nil {
				log.Println(err)
			}
			if !paperURL.IsAbs() {
				absURL := strings.Join([]string{readme.Path, link.Location}, "")
				link.Location = absURL
			}
			fmt.Println(link.Name)
			fmt.Println("\t", link.Location)
		}
	}
}
