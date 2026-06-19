package vsys_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vsys"
)

func ExampleGetCurrentPID() {
	pid := vsys.GetCurrentPID()
	fmt.Println(pid > 0)
	// Output: true
}

func ExampleEnvWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "APP_MODE" {
			return "test", true
		}
		return "", false
	})

	fmt.Println(vsys.EnvWithOptions("APP_MODE", lookup))
	fmt.Println(vsys.EnvWithOptions("MISSING", lookup))
	// Output:
	// test
	//
}

func ExampleEnvOrDefaultWithOptions() {
	lookup := vsys.WithEnvLookupFunc(func(key string) (string, bool) {
		if key == "PORT" {
			return "8080", true
		}
		return "", false
	})

	fmt.Println(vsys.EnvOrDefaultWithOptions("MISSING", "fallback", lookup))
	fmt.Println(vsys.EnvIntWithOptions("PORT", 0, lookup))
	// Output:
	// fallback
	// 8080
}

func ExampleGetCurrentPIDWithOptions() {
	pid := vsys.GetCurrentPIDWithOptions(vsys.WithPIDFunc(func() int {
		return 4242
	}))

	fmt.Println(pid)
	// Output: 4242
}

func ExampleNewGoInfoWithOptions() {
	info := vsys.NewGoInfoWithOptions(
		vsys.WithGoVersionFunc(func() string { return "go1.22.0" }),
		vsys.WithGoCompilerFunc(func() string { return "gc" }),
		vsys.WithGoRootFunc(func() string { return "/sdk/go" }),
		vsys.WithGoOSFunc(func() string { return "linux" }),
		vsys.WithGoArchFunc(func() string { return "amd64" }),
		vsys.WithGoNumCPUFunc(func() int { return 8 }),
		vsys.WithGoNumCgoCallFunc(func() int64 { return 0 }),
	)

	fmt.Println(info.Version, info.GOOS, info.NumCPU)
	// Output: go1.22.0 linux 8
}
