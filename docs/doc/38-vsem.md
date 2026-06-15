# vsem Quickstart

`vsem` provides counting semaphores with support for weights, context cancellation, non-blocking acquire attempts, and close semantics.

## Create semaphores and acquire/release permits

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem := vsem.New(2)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}
	fmt.Println(sem.Cap(), sem.Use())
	sem.Release(1)
	fmt.Println(sem.Use())
}
```

## Acquire permits without blocking

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem := vsem.New(1)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}
	fmt.Println(sem.TryAcquire(1))
	sem.Release(1)
	fmt.Println(sem.TryAcquire(1))
	sem.Release(1)
}
```

## Control wait timeouts with context

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem := vsem.New(1)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	err := sem.Acquire(ctx, 1)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
	sem.Release(1)
}
```

## Check errors and close semaphores

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/vsem"
)

func main() {
	sem, err := vsem.NewE(0)
	fmt.Println(sem == nil, errors.Is(err, vsem.ErrInvalidCapacity))

	active := vsem.New(1)
	active.Close()
	err = active.Acquire(context.Background(), 1)
	fmt.Println(errors.Is(err, vsem.ErrClosed))
}
```
