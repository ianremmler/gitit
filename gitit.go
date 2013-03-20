package gitit

import (
	"github.com/ianremmler/dgrl"
	"github.com/ianremmler/gitgo"

	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type GitIT struct {
	branchPath    string
	issueFilename string
}

func New() *GitIT {
	return &GitIT{
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
	n := IdToNum(id)
	if n < 0 {
		return id
	}
	return NumToId(n)
}

func (it *GitIT) IssueFilename() string {
	return it.issueFilename
}

func (it *GitIT) IdToBranch(id string) string {
	if id == "master" {
		return id
	}
	id = FormatId(id)
	if id == "" {
		return ""
	}
	return it.branchPath + id
}

func (it *GitIT) BranchToId(branch string) string {
	if branch == "master" {
		return branch
	}
	if strings.HasPrefix(branch, it.branchPath) {
		return branch[len(it.branchPath):]
	}
	return ""
}

func (it *GitIT) IssueIds() []string {
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

func (it *GitIT) ValidRepo() bool {
	repo := gitgo.New()
	_, err := repo.Run("status")
	return err == nil
}

func (it *GitIT) ValidIssue(id string) bool {
	repo := gitgo.New()
	_, err := repo.Run("show-ref", "-q", "--verify", "refs/heads/" + it.IdToBranch(id))
	return err == nil
}

func (it *GitIT) Init() error {
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

func (it *GitIT) MaxId() string {
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

func (it *GitIT) NewIssue() (string, error) {
	repo := gitgo.New()
	id := NextId(it.MaxId())
	_, err := repo.CheckoutNewBranch(it.IdToBranch(id))
	if err != nil {
		return "", err
	}
	return id, nil
}

func (it *GitIT) OpenIssue(id string) error {
	repo := gitgo.New()
	_, err := repo.Checkout(it.IdToBranch(id))
	return err
}

func (it *GitIT) CurIssue() (string, error) {
	repo := gitgo.New()
	branch, err := repo.CurBranch()
	if err != nil {
		return "", err
	}
	return it.BranchToId(branch), nil
}

func (it *GitIT) Cancel() error {
	repo := gitgo.New()
	_, err := repo.Reset("--hard")
	if err != nil {
		return err
	}
	_, err = repo.Checkout("master")
	return err
}

func (it *GitIT) SaveIssue() error {
	repo := gitgo.New()
	_, err := repo.Add(it.issueFilename)
	if err != nil {
		return err
	}
	_, err = repo.Commit("Updated issue.")
	return err
}

func (it *GitIT) AttachFile(filename string) error {
	repo := gitgo.New()
	_, err := repo.Add(filename)
	return err
}

func (it *GitIT) IssueText(id string) string {
	repo := gitgo.New()
	branch := ""
	if id != "" {
		branch = it.IdToBranch(id)
	}
	text, _ := repo.FileContents(branch, it.issueFilename)
	return text
}

func (it *GitIT) Blame(id string) string {
	repo := gitgo.New()
	branch := ""
	if id != "" {
		branch = it.IdToBranch(id)
	}
	text, _ := repo.Blame(branch, it.issueFilename)
	return text
}

func (it *GitIT) Value(id, key string) (string, bool) {
	return value(it.IssueText(id), key)
}

func (it *GitIT) WorkingValue(id, key string) (string, bool) {
	data, err := ioutil.ReadFile(it.issueFilename)
	if err != nil {
		return "", false
	}
	return value(string(data), key)
}

func value(text, key string) (string, bool) {
	parser := dgrl.NewParser()
	tree := parser.Parse(strings.NewReader(text))
	for _, node := range tree.Kids() {
		if leaf, ok := node.(*dgrl.Leaf); ok {
			if leaf.Key() == key {
				return leaf.Value(), true
			}
		}
	}
	return "", false
}

func (it *GitIT) SetWorkingValue(key, val string) bool {
	data, err := ioutil.ReadFile(it.issueFilename)
	if err != nil {
		return false
	}
	parser := dgrl.NewParser()
	tree := parser.Parse(strings.NewReader(string(data)))
	for _, node := range tree.Kids() {
		if leaf, ok := node.(*dgrl.Leaf); ok {
			if leaf.Key() == key {
				leaf.SetValue(val)
				issueFile, err := os.Create(it.issueFilename)
				if err != nil {
					return false
				}
				issueFile.WriteString(tree.String())
				issueFile.Close()
				return true
			}
		}
	}
	return false
}

func (it *GitIT) MatchingIssues(key, val string) []string {
	matches := []string{}
	for _, id := range it.IssueIds() {
		if it.IssueContains(id, key, val) {
			matches = append(matches, id)
		}
	}
	return matches
}

func (it *GitIT) IssueContains(id, key, val string) bool {
	if key == "" {
		return true
	}
	parser := dgrl.NewParser()
	tree := parser.Parse(strings.NewReader(it.IssueText(id)))
	for _, node := range tree.Kids() {
		if leaf, ok := node.(*dgrl.Leaf); ok {
			if leaf.Key() == key && (val == "" || val == leaf.Value()) {
				return true
			}
		}
	}
	return false
}
