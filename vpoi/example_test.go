package vpoi_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/vpoi"
)

func ExampleSheetNames() {
	names, err := vpoi.SheetNames("nonexistent.xlsx")
	fmt.Println(names)
	fmt.Println(err != nil)
	// Output:
	// []
	// true
}

func ExampleWriteRowsToBuffer() {
	rows := [][]string{
		{"Name", "Age"},
		{"Alice", "30"},
	}

	buf, err := vpoi.WriteRowsToBuffer("Users", rows)
	fmt.Println(buf.Len() > 0)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}

func ExampleReadRowsFromReader() {
	rows := [][]string{
		{"Name", "Age"},
		{"Alice", "30"},
	}
	buf, _ := vpoi.WriteRowsToBuffer("Users", rows)
	got, err := vpoi.ReadRowsFromReader(bytes.NewReader(buf.Bytes()))

	fmt.Println(got)
	fmt.Println(err)
	// Output:
	// [[Name Age] [Alice 30]]
	// <nil>
}

func ExampleReadRowsFromReader_withReadSheet() {
	buf, _ := vpoi.WriteRowsToBuffer("Reports", [][]string{{"Q1", "10"}})
	rows, err := vpoi.ReadRowsFromReader(bytes.NewReader(buf.Bytes()), vpoi.WithReadSheet("Reports"))

	fmt.Println(rows)
	fmt.Println(err)
	// Output:
	// [[Q1 10]]
	// <nil>
}

func ExampleWriteRowsToBuffer_emptySheetName() {
	buf, err := vpoi.WriteRowsToBuffer("", [][]string{{"Name"}})

	fmt.Println(buf == nil)
	fmt.Println(errors.Is(err, vpoi.ErrEmptySheetName))
	// Output:
	// true
	// true
}

func ExampleWithReadSheet() {
	fmt.Println(vpoi.WithReadSheet("Reports") != nil)
	// Output: true
}

func ExampleWithWriteSheet() {
	fmt.Println(vpoi.WithWriteSheet("Reports") != nil)
	// Output: true
}
