package main

import (
	"remmler.org/go/gitit.git"

	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	if len(os.Args) < 2 {
		log.Fatalln("You must specify a command")
	}

	it := gitit.New()

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
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
			status := it.Field(id, "status")
			summary := it.Field(id, "summary")
			priority := it.Field(id, "priority")
			fmt.Printf("%s %-8s %-8s %s\n", id, status, priority, summary)
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
	case "cur":
		fmt.Println(idStr(it, ""))
	default:
		log.Fatalln(cmd + " is not a valid command")
	}
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
