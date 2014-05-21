package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/ianremmler/gitit"
)

const usage = `usage:

spec: all | <id(s)> | with <key> [<val>]

it [help | usage]      Show usage
it state [<spec>]      Show issue state
it init                Initialize new issue tracker
it new                 Create new issue
it id [<spec>]         List ids, optionally filtering by key/value
it show [<spec>]       Show issue (default: current)
it open <id>           Open issue
it save                Save current issue
it close               Save any pending changes and close current issue
it cancel              Cancel any pending changes and close current issue
it edit [<id>]         Edit issue (default: current)
it blame [<id>]        Show 'git blame' for issue (default: current)
it attach [<file(s)>]  Attach file to current issue`

var (
	args = os.Args[1:]
	it   = gitit.New()
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("it: ")

	cmd := ""
	if len(args) > 0 {
		cmd = strings.ToLower(args[0])
		args = args[1:]
	}
	switch cmd {
	case "", "-h", "-help", "--help", "help", "-u", "-usage", "--usage", "usage":
		usageCmd()
	case "state":
		stateCmd()
	case "init":
		initCmd()
	case "new":
		newCmd()
	case "id":
		idCmd()
	case "show":
		showCmd()
	case "open":
		openCmd()
	case "save":
		saveCmd()
	case "close":
		closeCmd()
	case "cancel":
		cancelCmd()
	case "set":
		setCmd()
	case "blame":
		blameCmd()
	case "edit":
		editCmd()
	case "attach":
		attachCmd()
	default:
		log.Fatalln(cmd + " is not a valid command\n\n" + usage)
	}
}

func usageCmd() {
	fmt.Println(usage)
}

func initCmd() {
	if it.Init() != nil {
		log.Fatalln("init: Error initializing issue tracker")
	}
}

func newCmd() {
	verifyRepo()
	id, err := it.NewIssue()
	if err != nil {
		log.Fatalln("new: Error creating new issue")
	}
	fmt.Println(id)
}

func stateCmd() {
	verifyRepo()
	curId := it.CurrentIssue()
	for _, id := range specIds(args) {
		statusChar := ' '
		if id == curId {
			statusChar = '*'
			if it.Dirty() {
				statusChar = '!'
			}
		}
		fmt.Printf("%c %s\n", statusChar, stateSummary(id))
	}
}

func idCmd() {
	verifyRepo()
	for _, id := range specIds(args) {
		fmt.Println(id)
	}
}

func showCmd() {
	verifyRepo()
	ids := specIds(args)
	for _, id := range ids {
		verifyIssue(id)
	}
	fmt.Print(it.ToDgrl(ids))
}

func openCmd() {
	if len(args) == 0 {
		log.Fatalln("open: You must specify an issue to open")
	}
	verifyRepo()
	if it.Dirty() {
		log.Fatalln("open: Unsaved changes in currently open issue")
	}
	id := gitit.FormatId(args[0])
	err := it.OpenIssue(id)
	if err != nil {
		log.Fatalln("open: Error opening issue " + id)
	}
}

func saveCmd() {
	verifyRepo()
	it.SaveIssue()
}

func closeCmd() {
	verifyRepo()
	it.SaveIssue()
	it.Cancel()
}

func cancelCmd() {
	verifyRepo()
	it.Cancel()
}

func setCmd() {
	verifyRepo()
	if len(args) < 2 {
		log.Fatalln("set: You must specify a key and value")
	}
	if !it.SetWorkingValue(args[0], args[1]) {
		log.Fatalln("set: Error setting value")
	}
}

func blameCmd() {
	verifyRepo()
	id := it.CurrentIssue()
	if len(args) > 0 {
		id = args[0]
	}
	verifyIssue(id)
	fmt.Print(it.Blame(id))
}

func editCmd() {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		log.Fatalln("VISUAL or EDITOR environment variable must be set")
	}
	verifyRepo()
	id := it.CurrentIssue()
	isCur := true
	if len(args) > 0 {
		id = gitit.FormatId(args[0])
		isCur = false
	}
	verifyIssue(id)
	if !isCur {
		err := it.OpenIssue(id)
		if err != nil {
			log.Fatalln("edit: Unable to open issue " + id)
		}
	}
	cmd := exec.Command(editor, it.IssueFilename())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func attachCmd() {
	if len(args) == 0 {
		log.Fatalln("attach: You must specify a file to attach")
	}
	verifyRepo()
	for i := range args {
		if it.AttachFile(args[i]) != nil {
			log.Fatalln("attach: Error attaching " + args[i])
		}
	}
}

func stateSummary(id string) string {
	verifyIssue(id)
	id = gitit.FormatId(id)
	status, _ := it.Value(id, "status")
	typ, _ := it.Value(id, "type")
	priority, _ := it.Value(id, "priority")
	assigned, _ := it.Value(id, "assigned")
	summary, _ := it.Value(id, "summary")
	return fmt.Sprintf("%s %-7s %-7s %-7s %-7s %s", id, status, typ, priority, assigned, summary)
}

func verifyRepo() {
	if !it.ValidRepo() {
		log.Fatalln("Issue tracker repository not found")
	}
}

func verifyIssue(id string) {
	if !it.ValidIssue(id) {
		log.Fatalln(id + " is not a valid issue")
	}
}

func matchIds(kv []string) []string {
	key, val := "", ""
	if len(kv) > 0 {
		key = kv[0]
	}
	if len(kv) > 1 {
		val = kv[1]
	}
	return it.MatchingIssues(key, val)
}

func specIds(args []string) []string {
	ids := []string{}
	switch {
	case len(args) == 0:
		ids = append(ids, it.CurrentIssue())
	case args[0] == "with":
		ids = matchIds(args[1:])
	case args[0] == "all":
		ids = it.IssueIds()
	default:
		ids = args
	}
	return ids
}
