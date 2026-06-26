package vcsv_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vcsv"
)

func ExampleReadString() {
	records, _ := vcsv.ReadString("a,b,c\n1,2,3\n")
	fmt.Println(records)
	// Output: [[a b c] [1 2 3]]
}

func ExampleWriteString() {
	out, err := vcsv.WriteString([][]string{
		{"name", "age"},
		{"alice", "30"},
	})

	fmt.Print(out)
	fmt.Println(err)
	// Output:
	// name,age
	// alice,30
	// <nil>
}

func ExampleRecordsToMaps() {
	rows, err := vcsv.RecordsToMaps([][]string{
		{"name", "age"},
		{"alice", "30"},
	})

	fmt.Println(rows[0]["name"], rows[0]["age"])
	fmt.Println(err)
	// Output:
	// alice 30
	// <nil>
}

func ExampleReadStringMaps() {
	rows, err := vcsv.ReadStringMaps("name,age\nalice,30\n")

	fmt.Println(rows[0]["name"], rows[0]["age"])
	fmt.Println(err)
	// Output:
	// alice 30
	// <nil>
}

func ExampleWriteStringStructs() {
	type Person struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}

	out, err := vcsv.WriteStringStructs([]Person{{Name: "alice", Age: 30}})

	fmt.Print(out)
	fmt.Println(err)
	// Output:
	// name,age
	// alice,30
	// <nil>
}
