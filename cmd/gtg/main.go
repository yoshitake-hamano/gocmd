
package main

import (
	"bytes"
	"io"
	"os"
	"io/ioutil"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
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

type GitGraphJsPrinter struct {
	BaseBranchHistory *BranchHistory
	SuppressTag       bool
}

func JsVarString(name string) string {
	reg := regexp.MustCompile("[[:^alnum:]]")
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
	id := c.ID().String()
	subject := strings.TrimSpace(strings.SplitN(c.Message, "\n", 2)[0])

	reg := regexp.MustCompile("[\"'\\\\]")
	subject = reg.ReplaceAllString(subject, "")

	w.WriteString(fmt.Sprintf("%s.commit({sha1: \"%s\", message: \"%s\"});\n",
		jsBranch, id[0:7], subject))
}

func (g *GitGraphJsPrinter)printTag(w io.StringWriter, branch, tag string) {
	if g.SuppressTag {
		return
	}
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
		for _, t := range(node.TagNames) {
			g.printTag(buf, node.BranchName, t)
		}
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

func filterSimpleNodes(nodes []*Node, filter func(index int, nodes []*Node) bool) []*Node {
	var remaining []*Node
	for i, n := range nodes {
		// if branch base node, do not filter
		if len(n.ChildBranches) != 0 {
			remaining = append(remaining, n)
			for _, b := range(n.ChildBranches) {
				b.Nodes = filterSimpleNodes(b.Nodes, filter)
			}
			continue
		}

		if filter(i, nodes) != true {
			remaining = append(remaining, n)
		}
	}
	remaining = updateNodeNext(remaining)
	return remaining
}

func filterAllTagsTarget(index int, nodes []*Node) bool {
	if filterSimpleTarget(index, nodes) != true {
		return false
	}

	n := nodes[index]
	if len(n.TagNames) != 0 {
		return false
	}
	return true
}

func filterAllTags(bh *BranchHistory) {
	bh.Nodes = filterSimpleNodes(bh.Nodes, filterAllTagsTarget)
}

func filterSimpleTarget(index int, nodes []*Node) bool {
	// if first commit in branch, do not filter
	if index == 0 {
		return false
	}
	// if last commit in branch, do not filter
	if index == len(nodes) - 1 {
		return false
	}
	return true
}

func filterSimple(bh *BranchHistory) {
	bh.Nodes = filterSimpleNodes(bh.Nodes, filterSimpleTarget)
}

func openCurrentRepository() (*git.Repository, error) {
	dir, _ := filepath.Abs(".")
	for {
		repo, err := git.PlainOpen(dir)
		if err == nil {
			return repo, nil
		}

		tmp := filepath.Dir(dir)
		if dir == tmp {
			return nil, fmt.Errorf("error: not found git repository")
		}
		dir = tmp
	}
}

func createBranchHistory(repo *git.Repository) (*BranchHistory, error) {
	bite, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	
	var baseHistory *BranchHistory
	err = bite.ForEach(func(bref *plumbing.Reference) error {
		log.Printf("found branch: %s\n", bref.Name())
	
		bh, err := NewBranchHistory(repo, bref)
		if err != nil {
			return err
		}
		if baseHistory == nil {
			baseHistory = bh
		} else {
			baseHistory.Add(bh)
		}
	
		return nil
	})
	if err != nil {
		return nil, err
	}
	if baseHistory == nil {
		return nil, fmt.Errorf("no branch found")
	}
	return baseHistory, nil
}

func createBranchHistoryOrder(repo *git.Repository, branchOrder []string) (*BranchHistory, error) {
	bite, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	
	var historyMap map[string]*BranchHistory = map[string]*BranchHistory{}
	err = bite.ForEach(func(bref *plumbing.Reference) error {
		log.Printf("found branch: %s\n", bref.Name())
	
		bh, err := NewBranchHistory(repo, bref)
		if err != nil {
			return err
		}
		historyMap[bref.Name().String()] = bh
	
		return nil
	})
	if err != nil {
		return nil, err
	}

	var baseHistory *BranchHistory
	for _, branch := range(branchOrder) {
		bh, ok := historyMap[branch]
		if ! ok {
			log.Printf("branch not found: %s", branch)
			continue
		}
		if baseHistory == nil {
			baseHistory = bh
		} else {
			baseHistory.Add(bh)
		}
	}
	if baseHistory == nil {
		return nil, fmt.Errorf("no branch found")
	}
	return baseHistory, nil
}

var branchNames stringsFlag

// not supported: merge commit
func main() {
	var (
		filterMode  = flag.String("f", "simple", "filter mode(full, alltags, simple)")
		verbose     = flag.Bool("v", false, "verbose")
		suppressTag = flag.Bool("suppress_tag", false, "suppress tag commit")
	)
	flag.Var(&branchNames, "b", "branch order(multi setting)(ex. -b refs/heads/master -b refs/heads/develop)")
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		cmd := os.Args[0]
		fmt.Fprintf(o, "Usage of %s:\n", cmd)
		fmt.Fprintf(o, "  %s is a git graph outputter\n", cmd)
		flag.PrintDefaults()
		fmt.Fprintf(o, `
support:
  - only first parent(see --first-parent in git command)

not support:
  - merge commit

`)
		fmt.Fprintf(o, "example:\n")
		fmt.Fprintf(o, "  %s -f full\n", cmd)
		fmt.Fprintf(o, "  %s -f alltags\n", cmd)
		fmt.Fprintf(o, "  %s -f simple\n", cmd)
		fmt.Fprintf(o, "  %s -suppress_tag=true\n", cmd)
		fmt.Fprintf(o, "  %s -suppress_tsg=true -v -f simple -b refs/heads/master -b refs/heads/develop\n", cmd)
	}
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	repo, err := openCurrentRepository()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var baseHistory *BranchHistory
	if len(branchNames) == 0 {
		baseHistory, err = createBranchHistory(repo)
	} else {
		baseHistory, err = createBranchHistoryOrder(repo, branchNames)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	tite, err := repo.Tags()
	check(err)

	// Addn tag infomation
	tite.ForEach(func(tref *plumbing.Reference) error {
		log.Printf("found tag: %s\n", tref.Name())

		n := baseHistory.Find(tref.Hash())
		if n == nil {
			return nil
		}

		n.TagNames = append(n.TagNames, tref.Name().String())
		return nil
	})

	switch *filterMode {
	case "full":
	case "alltags":
		filterAllTags(baseHistory)
	case "simple":
		filterSimple(baseHistory)
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported filter mode: %s\n", *filterMode)
		os.Exit(1)
	}

	ggjp := &GitGraphJsPrinter{
		BaseBranchHistory: baseHistory,
		SuppressTag: *suppressTag}
	t := template.New("")
	_, err = t.Parse(indexTemplate)
	check(err)

	err = t.Execute(os.Stdout,
		struct {Body string} {
			Body: ggjp.String()})
	check(err)
}
