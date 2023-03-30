package main

import (
	"fmt"
	"os"

	"github.com/bluet/redt-agent/agent"
)

func main() {
	if len(os.Args) == 1 {
		err := agent.RunShowMetrics()
		if err != nil {
			fmt.Println("Error running one-shot mode:", err)
			os.Exit(1)
		}
	} else {
		switch os.Args[1] {
		case "sysup":
			autoYes := false
			if len(os.Args) > 2 && os.Args[2] == "-y" {
				autoYes = true
			}
			err := agent.RunSysup(autoYes)
			if err != nil {
				fmt.Println("Error performing system upgrade:", err)
				os.Exit(1)
			}
		case "-d":
			agent.RunDaemon()
		default:
			fmt.Println("Invalid argument. Usage:")
			fmt.Println("./redt-agent                (show system metrics)")
			fmt.Println("./redt-agent sysup          (system upgrade)")
			fmt.Println("./redt-agent sysup -y       (system upgrade with automatic confirmation)")
			fmt.Println("./redt-agent -d             (daemon mode)")
			os.Exit(1)
		}
	}
}
