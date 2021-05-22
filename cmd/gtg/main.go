
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
	buf := bytes.NewBuffer(make([]byte, 0))
	for i, node := range(bh.Nodes) {
		branchName := bh.BranchReference.Name()
		buf.WriteString(fmt.Sprintf(" [%s] [%03d] %s %s\n", branchName, i, node.Commit.Hash, node.Commit.Author.When))

		for _, bh := range(node.Branches) {
			buf.WriteString(bh.String())
		}
	}
	return buf.String()
}

func (bh *BranchHistory)Add(other *BranchHistory) error {
	for i, node := range(other.Nodes) {
		if bh.Nodes[i].Commit.Hash == node.Commit.Hash {
			continue
		}
		other.Nodes = other.Nodes[i:]
		bh.Nodes[i].Branches = append(bh.Nodes[i].Branches, other)
		return nil
	}
	// same branch history
	return nil
}

type GitGraphJsPrinter struct {
	BaseBranchHistory *BranchHistory
}

func main() {
	// todo: get branch name from argument
	// todo: get tag filter from argument

	repo, err := git.PlainOpen(".")
	// todo: if error has occured, should find parent directory
	check(err)

	bite, err := repo.Branches()
	check(err)

	var baseHistory *BranchHistory
	bite.ForEach(func(bref *plumbing.Reference) error {
		bh, err := NewBranchHistory(repo, bref)
		check(err)

		if baseHistory == nil {
			baseHistory = bh
		}
		baseHistory.Add(bh)

		return nil
	})
	fmt.Printf("%s\n", baseHistory)
}
