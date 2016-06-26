package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/google/go-github/github"
)

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

func RandomReadme(dir string) (string, error) {
	pwl := "papers-we-love"

	if SkipDir(dir) {
		return RandomReadme("/")
	}

	client := github.NewClient(nil)
	fc, dc, _, err := client.Repositories.GetContents(pwl, pwl, dir, nil)
	if err != nil {
		return "", err
	}

	if fc == nil {
		randInt, err := RandomInt(len(dc))
		if err != nil {
			return "", err
		}
		randDir := dc[randInt]
		randDirName := randDir.Name

		readmePath := strings.Join([]string{*randDirName, "README.md"}, "/")
		return RandomReadme(readmePath)
	}

	readme, err := fc.GetContent()
	if err != nil {
		return "", err
	}

	return readme, nil
}

func main() {
	readme, err := RandomReadme("/")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(readme)
}
