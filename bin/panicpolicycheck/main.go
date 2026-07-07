package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type violation struct {
	ruleID  string
	message string
}

type checker struct {
	root       string
	violations []violation
}

var allowedPanicPaths = map[string]struct{}{
	"internal/bloomfilter/bitset_bloomfilter.go": {},
	"internal/bloomfilter/filter.go":             {},
	"internal/cron/pattern.go":                   {},
	"internal/db/db.go":                          {},
	"internal/errx/exit.go":                      {},
	"internal/job/map.go":                        {},
	"internal/jwt/jwt.go":                        {},
	"internal/jwt/signer.go":                     {},
	"internal/jwt/signer_util.go":                {},
	"internal/maps/maps.go":                      {},
	"internal/obj/serialize.go":                  {},
	"internal/semaphore/semaphore.go":            {},
}

func main() {
	rootFlag := flag.String("root", "", "repository root")
	flag.Parse()

	root := strings.TrimSpace(*rootFlag)
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "PANIC POLICY VIOLATION: [PANIC_POLICY_INPUT_ERROR] cannot resolve working directory: %v\n", err)
			os.Exit(1)
		}
	}

	c := &checker{root: root}
	if err := c.run(); err != nil {
		c.addViolation("PANIC_POLICY_INPUT_ERROR", err.Error())
	}
	if len(c.violations) > 0 {
		for _, violation := range c.violations {
			fmt.Fprintf(os.Stderr, "PANIC POLICY VIOLATION: [%s] %s\n", violation.ruleID, violation.message)
		}
		os.Exit(1)
	}
	fmt.Println("panic policy is valid")
}

func (c *checker) run() error {
	var files []string
	err := filepath.WalkDir(c.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, path := range files {
		c.checkFile(path)
	}
	return nil
}

func (c *checker) checkFile(path string) {
	rel, err := filepath.Rel(c.root, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		c.addViolation("PANIC_POLICY_PARSE_ERROR", fmt.Sprintf("%s: cannot parse Go file: %v", rel, err))
		return
	}
	ast.Inspect(file, func(node ast.Node) bool {
		fn, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}
		if fn.Name != nil && (strings.HasPrefix(fn.Name.Name, "Must") || strings.HasPrefix(fn.Name.Name, "Panic")) {
			return false
		}
		if _, ok := allowedPanicPaths[rel]; ok {
			return false
		}
		ast.Inspect(fn.Body, func(child ast.Node) bool {
			call, ok := child.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if !ok || ident.Name != "panic" {
				return true
			}
			pos := fileSet.Position(call.Pos())
			c.addViolation(
				"PANIC_POLICY_PRODUCTION_PANIC",
				fmt.Sprintf("%s:%d: production panic is not allowed outside known compatibility or Must/Panic-style APIs", rel, pos.Line),
			)
			return true
		})
		return false
	})
}

func (c *checker) addViolation(ruleID, message string) {
	c.violations = append(c.violations, violation{ruleID: ruleID, message: message})
}
