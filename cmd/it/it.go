package main

import (
	"github.com/ianremmler/gitit"

	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const usage = `usage:

it help | usage          Show usage
it init                  Initialize new issue tracker
it new                   Create new issue
it list                  List issues
it show [<id>]           Show issue
it open <id>             Open issue
it save                  Save current issue
it close                 Save any pending changes and close current issue
it cancel                Cancel any pending changes and close current issue
it edit [<id>]           Edit issue
it find [<key> [<val>]]  Find issues with given key and value
it blame [<id>]          Show 'git blame' for issue
it edit [<id>]           Edit issue
it status                Show status of current issue
it attach | add          Attach file to current issue
it todgrl                Export issues to Doggerel format
it tojson                Export issues to JSON format`

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
	case "init":
		initCmd()
	case "new":
		newCmd()
	case "list":
		listCmd()
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
	case "find":
		findCmd()
	case "set":
		setCmd()
	case "blame":
		blameCmd()
	case "edit":
		editCmd()
	case "status":
		statusCmd()
	case "attach", "add":
		attachCmd()
	case "todgrl":
		dgrlCmd()
	case "tojson":
		jsonCmd()
	default:
		log.Fatalln(cmd + " is not a valid command")
	}
}

func usageCmd() {
	fmt.Println(usage)
}

func initCmd() {
	if it.Init() != nil {
		log.Fatalln("Error initializing issue tracker")
	}
}

func newCmd() {
	verifyRepo()
	id, err := it.NewIssue()
	if err != nil {
		log.Fatalln("Error creating new issue")
	}
	fmt.Println(idStr(id))
}

func listCmd() {
	verifyRepo()
	curId := it.CurrentIssue()
	for _, id := range it.IssueIds() {
		statusChar := ' '
		if id == curId {
			statusChar = '*'
		}
		fmt.Printf("%c %s\n", statusChar, issueStatus(id))
	}
}

func showCmd() {
	verifyRepo()
	id := ""
	if len(args) > 0 {
		id = args[0]
		verifyIssue(id)
		fmt.Printf("%s\n\n%s", idStr(id), it.IssueText(id))
	} else {
		fmt.Printf("%s\n\n%s", idStr(it.CurrentIssue()), it.WorkingIssueText())
	}
}

func openCmd() {
	if len(args) == 0 {
		log.Fatalln("You must specify an issue to open")
	}
	verifyRepo()
	id := gitit.FormatId(args[0])
	err := it.OpenIssue(id)
	if err != nil {
		log.Fatalln("Error opening issue " + id)
	}
	fmt.Println(idStr(id))
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

func findCmd() {
	verifyRepo()
	key, val := "", ""
	if len(args) > 0 {
		key = args[0]
	}
	if len(args) > 1 {
		val = args[1]
	}
	matches := it.MatchingIssues(key, val)
	for _, id := range matches {
		fmt.Println(id)
	}
}

func setCmd() {
	verifyRepo()
	if len(args) < 2 {
		log.Fatalln("You must specify a key and value")
	}
	if !it.SetWorkingValue(args[0], args[1]) {
		log.Fatalln("Error setting value")
	}
}

func blameCmd() {
	verifyRepo()
	id := ""
	if len(args) > 0 {
		id = args[0]
	} else {
		id = it.CurrentIssue()
	}
	verifyIssue(id)
	fmt.Printf("%s\n\n%s", idStr(id), it.Blame(id))
}

func editCmd() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		log.Fatalln("ERROR or VISUAL environment variable must be set")
	}
	verifyRepo()
	id := ""
	isCur := (len(args) == 0)
	if isCur {
		id = it.CurrentIssue()
	} else {
		id = gitit.FormatId(args[0])
	}
	verifyIssue(id)
	if !isCur {
		err := it.OpenIssue(id)
		if err != nil {
			log.Fatalln("Unable to open issue " + id)
		}
	}
	cmd := exec.Command(editor, it.IssueFilename())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func statusCmd() {
	id := ""
	if len(args) > 0 {
		id = args[0]
	} else {
		id = it.CurrentIssue()
	}
	verifyIssue(id)
	fmt.Println(issueStatus(id))
}

func attachCmd() {
	if len(args) == 0 {
		log.Fatalln("You must specify a file to attach")
	}
	verifyRepo()
	if it.AttachFile(args[0]) != nil {
		log.Fatalln("Error attaching " + args[0])
	}
}

func dgrlCmd() {
	fmt.Print(it.ToDgrl())
}

func jsonCmd() {
	fmt.Println(it.ToDgrl().ToJson())
}

func issueStatus(id string) string {
	id = gitit.FormatId(id)
	status, _ := it.Value(id, "status")
	summary, _ := it.Value(id, "summary")
	priority, _ := it.Value(id, "priority")
	return fmt.Sprintf("%s %-8s %-8s %s", id, status, priority, summary)
}

func idStr(id string) string {
	return "[" + gitit.FormatId(id) + "]"
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
