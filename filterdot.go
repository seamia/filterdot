// Copyright 2021 Seamia Corporation. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const (
	addParentDepth = 2 // how many parents to add to the "included" nodes
	space          = " "
	spaces         = " \t"
)

type (
	msi  = map[string]int
	msb  = map[string]bool
	msas = map[string][]string
)

var (
	split     = strings.Split
	trim      = strings.Trim
	replace   = strings.ReplaceAll
	alert     = fmt.Println
	lowercase = strings.ToLower
)

// cli args:
// app from to +12 -23 +323 -4342

func main() {
	from, to, inclusions, exclusions, noDups := fromArgs()

	data, err := ioutil.ReadFile(from)
	if err != nil {
		alert("failed to open file", from, ", due to:", err)
		return
	}

	lines := split(string(data), "\n")

	reverse := make(msas)
	direct := make(msas)
	names := make(map[string]string)

	for _, line := range lines {
		line = trim(replace(line, "\t", space), space)
		if connector, left, right := isConnector(line); connector {
			reverse[right] = append(reverse[right], left)
			direct[left] = append(direct[left], right)

		} else if strings.Contains(line, "[label=") {
			name := getBetween(line, "<name> ", "\"];")
			name = trim(split(name, "|")[0], spaces)
			names[getBetween(line, "", " [")] = name
		}
	}

	tree := make(msi)
	exclude := make(msb)

	for _, excl := range exclusions {
		exclude[excl] = true
	}

	if len(inclusions) > 0 {
		for _, incl := range inclusions {
			root := incl
			addChildren(root, exclude, direct, &tree)
			addParents(root, reverse, addParentDepth, &tree)
		}
	} else {
		for k, _ := range direct {
			tree[k]++
		}
		for k, _ := range reverse {
			tree[k]++
		}
		for _, excl := range exclusions {
			delete(tree, excl)
		}
	}

	target, err := os.Create(to)
	if err != nil {
		alert("failed to create file", to, ", due to:", err)
		return
	}
	defer target.Close()

	media := bufio.NewWriter(target)
	showSelectedOnly(lines, tree, media, noDups)
	media.Flush()
}

func addParents(from string, reverse msas, count int, store *msi) {
	if store == nil {
		panic("are you nuts?")
	}

	(*store)[from]++

	if count == 0 {
		return
	}

	for _, parent := range reverse[from] {
		if _, found := (*store)[parent]; !found {
			addParents(parent, reverse, count-1, store)
		}
	}
}

func addChildren(from string, exclude msb, direct msas, store *msi) {
	if store == nil || len(from) == 0 {
		panic("are you nuts?")
	}

	(*store)[from]++

	if exclude[from] {
		return
	}

	for _, child := range direct[from] {
		if _, found := (*store)[child]; !found {
			addChildren(child, exclude, direct, store)
		}
	}
}

func getBetween(from, left, right string) string {

	if len(left) > 0 {
		l := strings.Index(from, left)
		if l < 0 {
			return ""
		}
		from = from[l+len(left):]
	}

	if len(right) > 0 {
		r := strings.Index(from, right)
		if r < 0 {
			return ""
		}
		from = from[:r]
	}
	return trim(from, " \t\r\n")
}

func showSelectedOnly(lines []string, includes msi, media io.Writer, noDups bool) {
	already := make(map[string]int)

	for _, line := range lines {
		normalized := trim(replace(line, "\t", space), space)
		if connector, left, right := isConnector(normalized); connector {
			if includes[left] > 0 && includes[right] > 0 {
				marker := left + "::" + right // this ignores possible ports
				if !noDups || already[marker] == 0 {
					fmt.Fprintln(media, line)
					already[marker]++
				}
			}
		} else if strings.Contains(normalized, " [label=") {
			number := getBetween(normalized, "", " [")
			if includes[number] > 0 {
				fmt.Fprintln(media, line)
			}
		} else {
			fmt.Fprintln(media, line)
		}
	}
}

func fromArgs() (from string, to string, inclusions []string, exclusions []string, noDups bool) {
	if len(os.Args) < 4 {
		alert("not enough arguments")
		os.Exit(7)
	}
	from = os.Args[1]
	to = os.Args[2]

	for _, arg := range os.Args[3:] {
		if strings.HasPrefix(arg, "+") {
			inclusions = append(inclusions, arg[1:])
		} else if strings.HasPrefix(arg, "-") {
			exclusions = append(exclusions, arg[1:])
		} else {
			if lowercase(arg) == "nodups" {
				noDups = true
			} else {
				inclusions = append(inclusions, arg)
			}
		}
	}

	if list, err := fromFile("include.txt"); err == nil {
		inclusions = append(inclusions, list...)
	}
	if list, err := fromFile("exclude.txt"); err == nil {
		exclusions = append(exclusions, list...)
	}

	return
}

func fromFile(name string) ([]string, error) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, line := range split(string(data), "\n") {
		id := split(trim(line, spaces), space)[0]
		if len(id) > 0 {
			result = append(result, id)
		}
	}
	return result, nil
}

func isConnector(txt string) (bool, string, string) {
	line := trim(replace(txt, "\t", space), space)
	parts := split(trim(line, " \t;"), "->")

	if len(parts) == 2 {
		left := trim(split(parts[0], ":")[0], spaces)
		right := trim(removeAttributes(split(parts[1], ":")[0]), spaces)
		return true, left, right
	}

	return false, "", ""
}

func removeAttributes(txt string) string {
	for _, char := range []string{"[", ";", "/", "*"} {
		if i := strings.Index(txt, char); i > 0 {
			txt = txt[:i]
		}
	}
	return txt
}
