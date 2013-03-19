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
	if id == "master" {
		return id
	}
	return NumToId(IdToNum(id))
}

func (it *GitIt) IssueFilename() string {
	return it.issueFilename
}

func (it *GitIt) IdToBranch(id string) string {
	if id == "master" {
		return id
	}
	id = FormatId(id)
	if id == "" {
		return ""
	}
	return it.branchPath + id
}

func (it *GitIt) BranchToId(branch string) string {
	if branch == "master" {
		return branch
	}
	if strings.HasPrefix(branch, it.branchPath) {
		return branch[len(it.branchPath):]
	}
	return ""
}

func (it *GitIt) IssueIds() []string {
	repo := gitgo.New()
	issueIds := []string{}
	branches, _ := repo.Branches(it.branchPath)
	for _, branch := range branches {
		id := it.BranchToId(branch)
		if id != "master" && id != "" {
			issueIds = append(issueIds, id)
		}
	}
	return issueIds
}

func (it *GitIt) Init() error {
	repo := gitgo.New()
	_, err := repo.Init()
	if err != nil {
		return err
	}
	issueFile, err := os.Create(it.issueFilename)
	if err != nil {
		return err
	}
	issueFile.WriteString(defaultIssue.String())
	issueFile.Close()
	_, err = repo.Add(it.issueFilename)
	if err != nil {
		return err
	}
	_, err = repo.Commit("Issue repo initialized.")
	if err != nil {
		return err
	}
	return nil
}

func (it *GitIt) MaxId() string {
	issues := it.IssueIds()
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

func (it *GitIt) NewIssue() (string, error) {
	repo := gitgo.New()
	id := NextId(it.MaxId())
	_, err := repo.CheckoutNewBranch(it.IdToBranch(id))
	if err != nil {
		return "", err
	}
	return id, nil
}

func (it *GitIt) OpenIssue(id string) error {
	repo := gitgo.New()
	_, err := repo.Checkout(it.IdToBranch(id))
	return err
}

func (it *GitIt) CurIssue() (string, error) {
	repo := gitgo.New()
	branch, err := repo.CurBranch()
	if err != nil {
		return "", err
	}
	return it.BranchToId(branch), nil
}

func (it *GitIt) Cancel() error {
	repo := gitgo.New()
	_, err := repo.Reset("--hard")
	if err != nil {
		return err
	}
	_, err = repo.Checkout("master")
	return err
}

func (it *GitIt) SaveIssue() error {
	repo := gitgo.New()
	_, err := repo.Add(it.issueFilename)
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

func (it *GitIt) IssueText(id string) string {
	repo := gitgo.New()
	branch := ""
	if id != "" {
		branch = it.IdToBranch(id)
	}
	out, _ := repo.FileContents(branch, it.issueFilename)
	return out
}

func (it *GitIt) Blame(id string) (string, error) {
	repo := gitgo.New()
	branch := ""
	if id != "" {
		branch = it.IdToBranch(id)
	}
	return repo.Blame(branch, it.issueFilename)
}

func (it *GitIt) Field(id, key string) string {
	issueText := it.IssueText(id)
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

func (it *GitIt) MatchingIssues(key, val string) []string {
	matches := []string{}
	for _, id := range it.IssueIds() {
		if it.IssueContains(id, key, val) {
			matches = append(matches, id)
		}
	}
	return matches
}

func (it *GitIt) IssueContains(id, key, val string) bool {
	if key == "" {
		return true
	}
	issueText := it.IssueText(id)
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
