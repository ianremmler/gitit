package main

import (
	"remmler.org/go/gitit.git"

	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const usage = `usage:

it [help | usage]        Show usage
it init                  Initialize new issue tracker
it new                   Create new issue
it list                  List issues
it show [<id>]           Show issue
it open <id>             Open issue
it save                  Save issue
it cancel                Cancel any pending changes and close issue
it edit [<id>]           Edit issue
it find [<key> [<val>]]  Find issues with given key and value
it blame [<id>]          Show 'git blame' for issue
it edit [<id>]           Edit issue`

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
	case "cancel":
		cancelCmd()
	case "find":
		findCmd()
	case "blame":
		blameCmd()
	case "edit":
		editCmd()
	case "status":
		statusCmd()
	case "attach":
		attachCmd()
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
	for _, id := range it.IssueIds() {
		fmt.Println(issueStatus(id))
	}
}

func showCmd() {
	verifyRepo()
	id := ""
	if len(args) > 0 {
		id = gitit.FormatId(args[0])
	} else {
		id, _ = it.CurIssue()
	}
	verifyIssue(id)
	fmt.Println(idStr(id) + "\n")
	fmt.Print(it.IssueText(id))
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
	fmt.Println(idStr(""))
	if it.SaveIssue() != nil {
		log.Fatalln("Error saving issue")
	}
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

func blameCmd() {
	verifyRepo()
	id := ""
	if len(args) > 0 {
		id = gitit.FormatId(args[0])
	}
	verifyIssue(id)
	fmt.Println(idStr(id) + "\n")
	fmt.Print(it.Blame(id))
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
		id, _ = it.CurIssue()
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
	if len(args) == 0 {
		id, _ = it.CurIssue()
	} else {
		id = gitit.FormatId(args[0])
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

func issueStatus(id string) string {
	id = gitit.FormatId(id)
	status := it.Field(id, "status")
	summary := it.Field(id, "summary")
	priority := it.Field(id, "priority")
	return fmt.Sprintf("%s %-8s %-8s %s", id, status, priority, summary)
}

func idStr(id string) string {
	if id != "" {
		return "id: " + gitit.FormatId(id)
	}
	curId, _ := it.CurIssue()
	if curId != "" {
		return "id: " + curId
	}
	return "id: ?"
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
