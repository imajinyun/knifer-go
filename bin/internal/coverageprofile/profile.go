// Package coverageprofile merges statement blocks across caller test binaries.
package coverageprofile

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Block struct {
	File, Range string
	Statements  int
	Covered     bool
}

var location = regexp.MustCompile(`^(.+\.go):([0-9]+\.[0-9]+,[0-9]+\.[0-9]+)$`)

func Parse(reader io.Reader) ([]Block, error) {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return nil, fmt.Errorf("coverage profile is empty")
	}
	mode := scanner.Text()
	if mode != "mode: set" && mode != "mode: count" && mode != "mode: atomic" {
		return nil, fmt.Errorf("invalid coverage mode %q", mode)
	}
	blocks := map[string]Block{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("malformed coverage line %q", line)
		}
		m := location.FindStringSubmatch(fields[0])
		if len(m) != 3 {
			return nil, fmt.Errorf("invalid block location %q", fields[0])
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n < 0 {
			return nil, fmt.Errorf("invalid statement count %q", fields[1])
		}
		count, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid execution count %q", fields[2])
		}
		block := Block{File: m[1], Range: m[2], Statements: n, Covered: count > 0}
		if previous, ok := blocks[fields[0]]; ok {
			if previous.Statements != n {
				return nil, fmt.Errorf("conflicting statement counts for %s", fields[0])
			}
			block.Covered = block.Covered || previous.Covered
		}
		blocks[fields[0]] = block
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("coverage profile has no blocks")
	}
	keys := make([]string, 0, len(blocks))
	for key := range blocks {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]Block, 0, len(keys))
	for _, key := range keys {
		out = append(out, blocks[key])
	}
	return out, nil
}

func Write(writer io.Writer, blocks []Block) error {
	if _, err := fmt.Fprintln(writer, "mode: atomic"); err != nil {
		return err
	}
	for _, b := range blocks {
		count := 0
		if b.Covered {
			count = 1
		}
		if _, err := fmt.Fprintf(writer, "%s:%s %d %d\n", b.File, b.Range, b.Statements, count); err != nil {
			return err
		}
	}
	return nil
}
