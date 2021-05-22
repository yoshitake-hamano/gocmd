
package main

import (
	"fmt"
	"bytes"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Node struct {
	Commit   *object.Commit
	Branches []*BranchHistory
}

type BranchHistory struct {
	Repository      *git.Repository
	BranchReference *plumbing.Reference
	Nodes           []*Node
}

func CreateNodeSlice(c *object.Commit) ([]*Node, error) {
	nodes := make([]*Node, 0)
	ite := c
	for {
		n := Node{Commit: ite}
		nodes = append([]*Node{&n}, nodes...)
		var err error
		ite, err = ite.Parent(0)
		if err == object.ErrParentNotFound {
			return nodes, nil
		}
		if err != nil {
			return nodes, err
		}
	}
}

func NewBranchHistory(repo *git.Repository, bref *plumbing.Reference) (*BranchHistory, error) {
	c, err := repo.CommitObject(bref.Hash())
	if err != nil {
		return nil, err
	}

	nodes, err := CreateNodeSlice(c)
	if err != nil {
		return nil, err
	}
	return &BranchHistory{
		Repository: repo,
		BranchReference: bref,
		Nodes: nodes,
	}, nil
}

func (bh *BranchHistory)String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 10))
	
	buf.WriteString(fmt.Sprintf("%s\n", bh.BranchReference.Name()))
	for i, node := range(bh.Nodes) {
		buf.WriteString(fmt.Sprintf(" [%03d] %s %s\n", i, node.Commit.Hash, node.Commit.Author.When))
	}
	return buf.String()
}

func main() {
	repo, err := git.PlainOpen(".")
	check(err)

	bite, err := repo.Branches()
	check(err)
	
	bite.ForEach(func(bref *plumbing.Reference) error {
		bh, err := NewBranchHistory(repo, bref)
		check(err)

		fmt.Printf("%s\n", bh)
		return nil
	})
}
