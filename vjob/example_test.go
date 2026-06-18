package vjob_test

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vjob"
)

func ExampleNewBatch() {
	results := make([]string, 0)
	batch := vjob.NewBatch(func(_ context.Context, vals []int) (vjob.Merge, error) {
		for _, v := range vals {
			results = append(results, fmt.Sprintf("processed:%d", v))
		}
		return nil, nil
	}, []int{1, 2, 3})
	_ = vjob.Run(context.Background(), batch)
	fmt.Println(results)
	// Output: [processed:1 processed:2 processed:3]
}
