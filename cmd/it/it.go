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
	default:
		log.Fatalln(cmd + " is not a valid command")
	}
}

func usageCmd() {
	fmt.Println(usage)
}

func initCmd() {
	it.Init()
}

func newCmd() {
	id, err := it.NewIssue()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(idStr(it, id))
}

func listCmd() {
	for _, id := range it.IssueIds() {
		fmt.Println(issueStatus(it, id))
	}
}

func showCmd() {
	id := ""
	if len(args) > 0 {
		id = gitit.FormatId(args[0])
	}
	if !it.ValidIssue(id) {
		log.Fatalln(id + " is not a valid issue")
	}
	fmt.Println(idStr(it, id) + "\n")
	fmt.Print(it.IssueText(id))
}

func openCmd() {
	if len(args) == 0 {
		log.Fatalln("You must specify an issue to open")
	}
	id := args[0]
	fmt.Println(idStr(it, id))
	it.OpenIssue(id)
}

func saveCmd() {
	fmt.Println(idStr(it, ""))
	it.SaveIssue()
}

func cancelCmd() {
	it.Cancel()
}

func findCmd() {
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
	id := ""
	if len(args) > 0 {
		id = args[0]
	}
	fmt.Println(idStr(it, id) + "\n")
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
	if len(args) > 0 {
		id := gitit.FormatId(args[0])
		if !it.ValidIssue(id) {
			log.Fatalln(id + " is not a valid issue")
		}
		it.OpenIssue(id)
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
		id, _ = it.CurIssue()
	}
	id = gitit.FormatId(id)
	if !it.ValidIssue(id) {
		log.Fatalln(id + " is not a valid issue")
	}
	fmt.Println(issueStatus(it, id))
}

func issueStatus(it *gitit.GitIt, id string) string {
	id = gitit.FormatId(id)
	status := it.Field(id, "status")
	summary := it.Field(id, "summary")
	priority := it.Field(id, "priority")
	return fmt.Sprintf("%s %-8s %-8s %s", id, status, priority, summary)
}

func idStr(it *gitit.GitIt, id string) string {
	if id != "" {
		return "id: " + gitit.FormatId(id)
	}
	curId, _ := it.CurIssue()
	if curId != "" {
		return "id: " + curId
	}
	return "id: ?"
}
