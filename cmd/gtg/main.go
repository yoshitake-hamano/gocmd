
package main

import (
	"io"
	"os"
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

func (g *GitGraphJsPrinter)printNode(w io.StringWriter, node *Node) {
	g.printCommit(w, node.BranchName, node.Commit)
	for _, b := range(node.ChildBranches) {
		newBranchName := b.BranchReference.Name().String()
		g.printBranch(w, node.BranchName, newBranchName)
	}
}

func (g *GitGraphJsPrinter)String() string {
	buf := bytes.NewBuffer(make([]byte, 0))
	baseBranch := g.BaseBranchHistory.BranchReference.Name().String()
	g.printFirstBranch(buf, baseBranch)

	for _, node := range g.BaseBranchHistory.Nodes {
		g.printNode(buf, node)
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

// not supported: merge commit
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


	ggjp := &GitGraphJsPrinter{baseHistory}
	t := template.New("index.html")
	_, err = t.Parse(indexTemplate)
	check(err)

	err = t.Execute(os.Stdout,
		struct {Body string} {
			Body: ggjp.String()})
	check(err)
}
