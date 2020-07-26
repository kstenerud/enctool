// Copyright 2020 Karl Stenerud
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package main

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/kstenerud/go-concise-encoding/debug"
)

func printUsage(cmd Command) {
	if cmd == nil {
		fmt.Printf("Usage: %v <command> [options]\n", os.Args[0])
		fmt.Printf("Available commands:\n")
		keys := make([]string, 0, len(commands))
		for k := range commands {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			cmd := commands[k]
			fmt.Printf("    %-10v: %v\n", k, cmd.Description())
		}
	} else {
		fmt.Println(cmd.Usage())
	}
}

func printUsageExit(cmd Command) {
	printUsage(cmd)
	os.Exit(1)
}

func printHelpExit() {
	printUsage(nil)
	os.Exit(0)
}

func main() {
	// debug.DebugOptions.PassThroughPanics = true
	args := os.Args
	if len(args) < 2 {
		printUsageExit(nil)
	}

	cmdName := args[1]
	if cmdName == "-help" || cmdName == "-h" || cmdName == "-?" {
		printHelpExit()
	}

	cmd := getCommand(cmdName)
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "%v: Unknown subcommand\n", cmdName)
		printUsageExit(cmd)
	}

	if err := cmd.Init(args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Stderr.Sync()
		if errors.Is(err, UsageError) {
			printUsageExit(cmd)
		}
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Stderr.Sync()
		if errors.Is(err, UsageError) {
			printUsageExit(cmd)
		}
		os.Exit(1)
	}
}
