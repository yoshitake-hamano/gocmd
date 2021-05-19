
package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	repo, err := git.PlainOpen(".")
	check(err)

	cite, err := repo.Log(&git.LogOptions{
		All: true,
	})
	check(err)

	cite.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
}
