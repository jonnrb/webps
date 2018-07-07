package main

import (
	"fmt"
	"os"

	"go.jonnrb.io/webps/cmd/webps-backend/runner"
	"go.jonnrb.io/webps/cmd/webps-frontend/runner"
	"go.jonnrb.io/webps/cmd/webps-keygen/runner"
)

func usage() {
	fmt.Println("usage: quay.io/jonnrb/webps frontend|backend|keygen ...")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	cmd, cmdArgs := os.Args[1], os.Args[2:]
	os.Args = append([]string{"quay.io/jonnrb/webps " + cmd}, cmdArgs...)

	var run func()
	switch cmd {
	case "frontend":
		run = frontend.Run
	case "backend":
		run = backend.Run
	case "keygen":
		run = keygen.Run
	default:
		run = usage
	}

	run()
}
