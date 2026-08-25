package vcsv_test

import (
	"bytes"
	"fmt"
	"strings"

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

func ExampleRead() {
	records, err := vcsv.Read(strings.NewReader("a,b\n1,2\n"))
	fmt.Println(records)
	fmt.Println(err)
	// Output:
	// [[a b] [1 2]]
	// <nil>
}

func ExampleReadMaps() {
	rows, err := vcsv.ReadMaps(strings.NewReader("name,age\nalice,30\n"))
	fmt.Println(rows[0]["name"], rows[0]["age"])
	fmt.Println(err)
	// Output:
	// alice 30
	// <nil>
}

func ExampleForEach() {
	var seen []string
	err := vcsv.ForEach(strings.NewReader("a\nb\n"), func(record []string) error {
		seen = append(seen, record[0])
		return nil
	})
	fmt.Println(strings.Join(seen, ""))
	fmt.Println(err)
	// Output:
	// ab
	// <nil>
}

func ExampleMapsToRecords() {
	records := vcsv.MapsToRecords([]string{"name", "age"}, []map[string]string{
		{"name": "alice", "age": "30"},
	})
	fmt.Println(records)
	// Output: [[name age] [alice 30]]
}

func ExampleStructsToRecords() {
	type Person struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}
	records, err := vcsv.StructsToRecords([]Person{{Name: "alice", Age: 30}})
	fmt.Println(records)
	fmt.Println(err)
	// Output:
	// [[name age] [alice 30]]
	// <nil>
}

func ExampleWrite() {
	var b bytes.Buffer
	err := vcsv.Write(&b, [][]string{{"name"}, {"alice"}})
	fmt.Print(b.String())
	fmt.Println(err)
	// Output:
	// name
	// alice
	// <nil>
}

func ExampleWriteMaps() {
	var b bytes.Buffer
	err := vcsv.WriteMaps(&b, []string{"name"}, []map[string]string{{"name": "alice"}})
	fmt.Print(b.String())
	fmt.Println(err)
	// Output:
	// name
	// alice
	// <nil>
}

func ExampleWriteStringMaps() {
	out, err := vcsv.WriteStringMaps([]string{"name"}, []map[string]string{{"name": "alice"}})
	fmt.Print(out)
	fmt.Println(err)
	// Output:
	// name
	// alice
	// <nil>
}

func ExampleWriteStructs() {
	type Person struct {
		Name string `csv:"name"`
		Age  int    `csv:"age"`
	}
	var b bytes.Buffer
	err := vcsv.WriteStructs(&b, []Person{{Name: "alice", Age: 30}})
	fmt.Print(b.String())
	fmt.Println(err)
	// Output:
	// name,age
	// alice,30
	// <nil>
}

func ExampleWithComma() {
	records, err := vcsv.ReadString("a;b\n1;2\n", vcsv.WithComma(';'))
	fmt.Println(records)
	fmt.Println(err)
	// Output:
	// [[a b] [1 2]]
	// <nil>
}
