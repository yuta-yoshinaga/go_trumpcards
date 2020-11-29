package main

import (
	"flag"
	"fmt"
	"go_trumpcards/frameworks_drivers/ui"
	"go_trumpcards/frameworks_drivers/web"
	"log"
	"strings"
)

func main() {
	flag.Parse()
	switch strings.ToLower(flag.Arg(0)) {
	case "cui":
		cui := ui.NewBlackJackCui()
		cui.Exec()
	case "web":
		web := web.NewBlackJackWeb()
		web.Exec()
	default:
		log.Fatal(fmt.Errorf("Error: param not found %s", flag.Arg(0)))
	}
}
