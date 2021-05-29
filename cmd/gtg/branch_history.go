package main

import (
	"bytes"
	"fmt"
	"log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Node struct {
	BranchName      string
	Commit          *object.Commit
	ChildBranches   []*BranchHistory
	TagNames        []string
	Next            *Node
}

type BranchHistory struct {
	Repository      *git.Repository
	BranchReference *plumbing.Reference
	Nodes           []*Node
}

func CreateNodeSlice(branch string, c *object.Commit) ([]*Node, error) {
	nodes := make([]*Node, 0)
	ite := c
	for {
		n := Node{BranchName: branch, Commit: ite, TagNames: []string{}}
		if len(nodes) != 0 {
			n.Next = nodes[0]
		}
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

	nodes, err := CreateNodeSlice(bref.Name().String(), c)
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

		for _, bh := range(node.ChildBranches) {
			buf.WriteString(bh.String())
		}
	}
	return buf.String()
}

func (bh *BranchHistory)Add(other *BranchHistory) error {
	prevBaseNode := bh.Nodes[0]
	baseNode := bh.Nodes[0]
        OUTER:
	for i, n := range(other.Nodes) {
		if baseNode.Commit.Hash == n.Commit.Hash {
			prevBaseNode = baseNode
			baseNode     = baseNode.Next
			continue
		}
		for _, branch := range(prevBaseNode.ChildBranches) {
			if branch.Nodes[0].Commit.Hash == n.Commit.Hash {
				prevBaseNode = branch.Nodes[0]
				baseNode     = branch.Nodes[0].Next
				continue OUTER
			}
		}
		other.Nodes = other.Nodes[i:]
		prevBaseNode.ChildBranches = append(prevBaseNode.ChildBranches, other)
		return nil
	}
	log.Printf("same or independent branch: %s %s\n",
		bh.BranchReference.Name().String(),
		other.BranchReference.Name().String())
	return nil
}

func (bh *BranchHistory)Find(hash plumbing.Hash) *Node {
	for _, n := range(bh.Nodes) {
		if hash == n.Commit.Hash {
			return n
		}

		for _, b := range(n.ChildBranches) {
			cn := b.Find(hash)
			if cn != nil {
				return cn
			}
		}
	}
	return nil
}
