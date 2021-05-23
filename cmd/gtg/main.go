
package main

import (
	"io"
	"os"
	"flag"
	"fmt"
	"regexp"
	"bytes"
	"text/template"
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
	BranchName      string
	Commit          *object.Commit
	ChildBranches   []*BranchHistory
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
		n := Node{BranchName: branch, Commit: ite}
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
	for i, node := range(other.Nodes) {
		if bh.Nodes[i].Commit.Hash == node.Commit.Hash {
			continue
		}
		other.Nodes = other.Nodes[i:]
		bh.Nodes[i-1].ChildBranches = append(bh.Nodes[i-1].ChildBranches, other)
		return nil
	}
	// same branch history
	return nil
}

type GitGraphJsPrinter struct {
	BaseBranchHistory *BranchHistory
}

func JsVarString(name string) string {
	reg := regexp.MustCompile("[[:^alpha:]]")
	return fmt.Sprintf("_%s", reg.ReplaceAllString(name, "_"))
}

func (g *GitGraphJsPrinter)printFirstBranch(w io.StringWriter, newBranch string) {
	jsNewBranch := JsVarString(newBranch)
	w.WriteString(fmt.Sprintf("var %s = %s.branch(\"%s\");\n", jsNewBranch, "gitgraph", newBranch))
}

func (g *GitGraphJsPrinter)printBranch(w io.StringWriter, baseBranch, newBranch string) {
	jsBaseBranch := JsVarString(baseBranch)
	jsNewBranch := JsVarString(newBranch)
	w.WriteString(fmt.Sprintf("var %s = %s.branch(\"%s\");\n", jsNewBranch, jsBaseBranch, newBranch))
}

func (g *GitGraphJsPrinter)printCommit(w io.StringWriter, branch string, c *object.Commit) {
	jsBranch := JsVarString(branch)
	w.WriteString(fmt.Sprintf("%s.commit(\"%s\");\n", jsBranch, c.ID()))
}

func (g *GitGraphJsPrinter)printTag(w io.StringWriter, branch, tag string) {
	jsBranch := JsVarString(branch)
	w.WriteString(fmt.Sprintf("%s.tag(\"%s\");\n", jsBranch, tag))
}

func removeOldestNode(nodes []*Node) (*Node, []*Node) {
	if len(nodes) == 0 {
		return nil, nodes
	}
	oldestIndex := 0
	for i, n := range(nodes) {
		oldest := nodes[oldestIndex].Commit.Author.When
		when   := n.Commit.Author.When
		if oldest.After(when) {
			oldestIndex = i
		}
	}
	n := nodes[oldestIndex]
	return n, append(nodes[:oldestIndex], nodes[oldestIndex+1:]...)
}

func (g *GitGraphJsPrinter)String() string {
	buf := bytes.NewBuffer(make([]byte, 0))
	baseBranch := g.BaseBranchHistory.BranchReference.Name().String()
	g.printFirstBranch(buf, baseBranch)

	if len(g.BaseBranchHistory.Nodes) == 0 {
		return buf.String()
	}
	nextNodes := []*Node{g.BaseBranchHistory.Nodes[0]}

	for {
		if len(nextNodes) == 0 {
			break
		}
		var node *Node
		node, nextNodes = removeOldestNode(nextNodes)
		if node.Next != nil {
			nextNodes = append(nextNodes, node.Next)
		}

		g.printCommit(buf, node.BranchName, node.Commit)
		for _, b := range(node.ChildBranches) {
			newBranchName := b.BranchReference.Name().String()
			g.printBranch(buf, node.BranchName, newBranchName)

			nextNodes = append(nextNodes, b.Nodes[0])
		}
	}
	return buf.String()
}

const indexTemplate = `
<html>
<head>
<script src="https://cdnjs.cloudflare.com/ajax/libs/gitgraph.js/1.8.3/gitgraph.min.js"></script>
<link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/gitgraph.js/1.8.3/gitgraph.min.css" />
</head>

<body>
    <canvas id="gitGraph"></canvas>
</body>

<script>
var gitgraph = new GitGraph({
    template: "metro",
    orientation: "horizontal",
    mode: "compact",
    elementId: "gitGraph"
});
{{.Body}}
</script>
</html>
`

func filterAllTags(bh *BranchHistory) {
}

func updateNodeNext(nodes []*Node) []*Node {
	for i, n := range nodes {
		if i == len(nodes) - 1 {
			n.Next = nil
		} else {
			n.Next = nodes[i+1]
		}
	}
	return nodes
}

func filterSimpleNodes(nodes []*Node) []*Node {
	var remaining []*Node
	for i, n := range nodes {
		// if first commit in branch, do not filter
		if i == 0 {
			remaining = append(remaining, n)
			continue
		}
		// if last commit in branch, do not filter
		if i == len(nodes) - 1 {
			remaining = append(remaining, n)
			continue
		}
		if len(n.ChildBranches) != 0 {
			remaining = append(remaining, n)
			for _, b := range(n.ChildBranches) {
				b.Nodes = filterSimpleNodes(b.Nodes)
			}
			continue
		}
	}
	remaining = updateNodeNext(remaining)
	return remaining
}

func filterSimple(bh *BranchHistory) {
	bh.Nodes = filterSimpleNodes(bh.Nodes)
}

// not supported: merge commit
func main() {
	var (
		filterMode  = flag.String("f", "", "filter mode(full, alltags, simple)")
	)
	flag.Parse()
	// todo: get branch name from argument

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

	switch *filterMode {
	case "full":
	case "alltags":
		filterAllTags(baseHistory)
	case "simple":
		filterSimple(baseHistory)
	default:
		check(fmt.Errorf("unsupported filter mode: %s", *filterMode))
	}

	ggjp := &GitGraphJsPrinter{baseHistory}
	t := template.New("")
	_, err = t.Parse(indexTemplate)
	check(err)

	err = t.Execute(os.Stdout,
		struct {Body string} {
			Body: ggjp.String()})
	check(err)
}
