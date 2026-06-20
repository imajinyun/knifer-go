# vjob Quickstart

`vjob` provides sliceable task scheduling helpers that split slices, ranges, or map keys into batches and run merge callbacks in order after shards succeed.

## Run range tasks

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vjob"
)

func main() {
	job := vjob.NewSlice(func(ctx context.Context, start, end int) (vjob.Merge, error) {
		fmt.Println("run", start, end)
		return func() error {
			fmt.Println("merge", start, end)
			return nil
		}, nil
	}, 10)

	err := vjob.RunWith(context.Background(), job, vjob.Options{
		BatchSize:      3,
		MaxConcurrency: 2,
	})
	if err != nil {
		panic(err)
	}
}
```

## Process typed slices

```go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/imajinyun/go-knifer/vjob"
)

func main() {
	values := []int{1, 2, 3, 4}
	var mu sync.Mutex
	total := 0

	job := vjob.NewBatch(func(ctx context.Context, batch []int) (vjob.Merge, error) {
		sum := 0
		for _, v := range batch {
			sum += v
		}
		return func() error {
			mu.Lock()
			defer mu.Unlock()
			total += sum
			return nil
		}, nil
	}, values).WithBatchSize(2).WithMaxConcurrency(2)

	if err := vjob.Run(context.Background(), job); err != nil {
		panic(err)
	}
	fmt.Println(total)
}
```

## Process slice elements one by one

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vjob"
)

func main() {
	items := []string{"go", "knifer", "job"}
	job := vjob.NewBatchSingle(func(ctx context.Context, item string) (vjob.Merge, error) {
		upper := len(item)
		return func() error {
			fmt.Println(item, upper)
			return nil
		}, nil
	}, items).WithMaxConcurrency(3)

	if err := vjob.Run(context.Background(), job); err != nil {
		panic(err)
	}
}
```

## Iterate map keys

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vjob"
)

func main() {
	scores := map[string]int{"alice": 90, "bob": 80}
	job := vjob.NewMapKeys(func(ctx context.Context, name string) (vjob.Merge, error) {
		return func() error {
			fmt.Println(name, scores[name])
			return nil
		}, nil
	}, scores).WithBatchSize(1)

	if err := vjob.Run(context.Background(), job); err != nil {
		panic(err)
	}
}
```
