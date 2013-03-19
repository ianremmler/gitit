package main

import (
	"remmler.org/go/gitit.git"

	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const usage = `it usage

it                       Show status of current issue
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

func main() {
	log.SetFlags(0)
	log.SetPrefix("it: ")

	it := gitit.New()

	cmd := ""
	if len(os.Args) >= 2 {
		cmd = strings.ToLower(os.Args[1])
	}
	switch cmd {
	case "", "-h", "-help", "--help", "help", "usage":
		fmt.Println(usage)
	case "init":
		it.Init()
	case "new":
		id, err := it.NewIssue()
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(idStr(it, id))
	case "list":
		for _, id := range it.IssueIds() {
			fmt.Println(issueStatus(it, id))
		}
	case "show":
		id := ""
		if len(os.Args) > 2 {
			id = os.Args[2]
		}
		fmt.Println(idStr(it, id) + "\n")
		fmt.Print(it.IssueText(id))
	case "open":
		if len(os.Args) < 3 {
			log.Fatalln("You must specify an issue to open")
		}
		id := os.Args[2]
		fmt.Println(idStr(it, id))
		it.OpenIssue(id)
	case "save":
		fmt.Println(idStr(it, ""))
		it.SaveIssue()
	case "cancel":
		it.Cancel()
	case "find":
		key, val := "", ""
		if len(os.Args) > 2 {
			key = os.Args[2]
		}
		if len(os.Args) > 3 {
			val = os.Args[3]
		}
		matches := it.MatchingIssues(key, val)
		for _, id := range matches {
			fmt.Println(id)
		}
	case "blame":
		id := ""
		if len(os.Args) > 2 {
			id = os.Args[2]
		}
		fmt.Println(idStr(it, id) + "\n")
		fmt.Print(it.Blame(id))
	case "edit":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			log.Fatalln("ERROR or VISUAL environment variable must be set")
		}
		if len(os.Args) > 2 {
			it.OpenIssue(os.Args[2])
		}
		cmd := exec.Command(editor, it.IssueFilename())
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
		}
	case "status":
		id := ""
		if len(os.Args) > 2 {
			id = os.Args[2]
		} else {
		    id, _ = it.CurIssue()
		}
		if id != "" {
			fmt.Println(issueStatus(it, id))
		}
	default:
		log.Fatalln(cmd + " is not a valid command")
	}
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
