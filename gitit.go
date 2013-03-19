package gitit

import (
	"github.com/ianremmler/dgrl"
	"remmler.org/go/gitgo.git"

	"fmt"
	"os"
	"strconv"
	"strings"
)

type GitIt struct {
	branchPath    string
	issueFilename string
}

func New() *GitIt {
	return &GitIt{
		branchPath:    "issue/",
		issueFilename: "issue",
	}
}

func NumToId(n int) string {
	return fmt.Sprintf("%04d", n)
}

func IdToNum(id string) int {
	n, err := strconv.Atoi(id)
	if err != nil {
		return -1
	}
	return n
}

func FormatId(id string) string {
	return NumToId(IdToNum(id))
}

func (g *GitIt) IssueFilename() string {
	return g.issueFilename
}

func (g *GitIt) IdToBranch(id string) string {
	id = FormatId(id)
	if id == "" {
		return ""
	}
	return g.branchPath + id
}

func (g *GitIt) BranchToId(branch string) string {
	if strings.HasPrefix(branch, g.branchPath) {
		return branch[len(g.branchPath):]
	}
	return ""
}

func (g *GitIt) IssueIds() []string {
	repo := gitgo.New()
	issueIds := []string{}
	branches, _ := repo.Branches(g.branchPath)
	for _, branch := range branches {
		if strings.HasPrefix(branch, g.branchPath) {
			id := branch[len(g.branchPath):]
			issueIds = append(issueIds, id)
		}
	}
	return issueIds
}

func (g *GitIt) Init() error {
	repo := gitgo.New()
	_, err := repo.Init()
	if err != nil {
		return err
	}
	issueFile, err := os.Create(g.issueFilename)
	if err != nil {
		return err
	}
	issueFile.WriteString(defaultIssue.String())
	issueFile.Close()
	_, err = repo.Add(g.issueFilename)
	if err != nil {
		return err
	}
	_, err = repo.Commit("Issue repo initialized.")
	if err != nil {
		return err
	}
	return nil
}

func (g *GitIt) MaxId() string {
	issues := g.IssueIds()
	if len(issues) == 0 {
		return NumToId(0)
	}
	max := 0
	for _, id := range issues {
		n := IdToNum(id)
		if n > max {
			max = n
		}
	}
	return NumToId(max)
}

func NextId(id string) string {
	return NumToId(IdToNum(id) + 1)
}

func (g *GitIt) NewIssue() (string, error) {
	repo := gitgo.New()
	id := NextId(g.MaxId())
	_, err := repo.CheckoutNewBranch(g.IdToBranch(id))
	if err != nil {
		return "", err
	}
	return id, nil
}

func (g *GitIt) OpenIssue(id string) error {
	repo := gitgo.New()
	_, err := repo.Checkout(g.IdToBranch(id))
	return err
}

func (g *GitIt) CurIssue() (string, error) {
	repo := gitgo.New()
	branch, err := repo.CurBranch()
	if err != nil {
		return "", err
	}
	return g.BranchToId(branch), nil
}

func (g *GitIt) Cancel() error {
	repo := gitgo.New()
	_, err := repo.Reset("--hard")
	if err != nil {
		return err
	}
	_, err = repo.Checkout("master")
	return err
}

func (g *GitIt) SaveIssue() error {
	repo := gitgo.New()
	_, err := repo.Add(g.issueFilename)
	if err != nil {
		return err
	}
	_, err = repo.Commit("Updated issue.")
	if err != nil {
		return err
	}
	_, err = repo.Checkout("master")
	return err
}

func (g *GitIt) IssueText(id string) string {
	repo := gitgo.New()
	branch := ""
	if id != "" {
		branch = g.IdToBranch(id)
	}
	out, _ := repo.FileContents(branch, g.issueFilename)
	return out
}

func (g *GitIt) Blame(id string) (string, error) {
	repo := gitgo.New()
	branch := ""
	if id != "" {
		branch = g.IdToBranch(id)
	}
	return repo.Blame(branch, g.issueFilename)
}

func (g *GitIt) Field(id, key string) string {
	issueText := g.IssueText(id)
	parser := dgrl.NewParser()
	tree := parser.Parse(strings.NewReader(issueText))
	for _, node := range tree.Kids() {
		if leaf, ok := node.(*dgrl.Leaf); ok {
			if leaf.Key() == key {
				return leaf.Value()
			}
		}
	}
	return ""
}

func (g *GitIt) MatchingIssues(key, val string) []string {
	matches := []string{}
	for _, id := range g.IssueIds() {
		if g.IssueContains(id, key, val) {
			matches = append(matches, id)
		}
	}
	return matches
}

func (g *GitIt) IssueContains(id, key, val string) bool {
	if key == "" {
		return true
	}
	issueText := g.IssueText(id)
	parser := dgrl.NewParser()
	tree := parser.Parse(strings.NewReader(issueText))
	for _, node := range tree.Kids() {
		if leaf, ok := node.(*dgrl.Leaf); ok {
			if leaf.Key() == key && (val == "" || val == leaf.Value()) {
				return true
			}
		}
	}
	return false
}
