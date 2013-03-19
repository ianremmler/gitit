package gitit

import (
	"github.com/ianremmler/dgrl"
)

var defaultIssue *dgrl.Branch

func init() {
	defaultIssue = dgrl.NewRoot()
	defaultIssue.Append(dgrl.NewLeaf("summary", ""))
	defaultIssue.Append(dgrl.NewLeaf("type", ""))
	defaultIssue.Append(dgrl.NewLeaf("status", ""))
	defaultIssue.Append(dgrl.NewLeaf("assigned", ""))
	defaultIssue.Append(dgrl.NewLongLeaf("description", ""))
}
