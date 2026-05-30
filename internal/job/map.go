package job

import (
	"context"
	"fmt"
	"reflect"
)

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// NewMapKeys creates a single-item job over typed map keys.
func NewMapKeys[K comparable, V any](run func(context.Context, K) (Merge, error), m map[K]V) *Batch[K] {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return NewBatchSingle(run, keys)
}

// NewMap creates a single-item job over map keys.
// The run function must accept context.Context and one key, and return (Merge, error).
func NewMap(run any, m any) *Slice {
	f := reflect.ValueOf(run)
	if f.Kind() != reflect.Func {
		panic(fmt.Errorf("job run must be a func, got %T", run))
	}
	if f.Type().NumIn() != 2 || !f.Type().In(0).Implements(contextType) || f.Type().NumOut() != 2 {
		panic(fmt.Errorf("job run must use func(context.Context, key) (job.Merge, error), got %s", f.Type()))
	}
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	mergeType := reflect.TypeOf(Merge(nil))
	if !f.Type().Out(0).AssignableTo(mergeType) || !f.Type().Out(1).Implements(errorType) {
		panic(fmt.Errorf("job run must return (job.Merge, error), got %s", f.Type()))
	}

	val := reflect.ValueOf(m)
	if !val.IsValid() || val.Kind() != reflect.Map {
		panic(fmt.Errorf("job data must be a map, got %T", m))
	}
	if !val.Type().Key().AssignableTo(f.Type().In(1)) {
		panic(fmt.Errorf("job map key %s is not assignable to run arg %s", val.Type().Key(), f.Type().In(1)))
	}

	keys := val.MapKeys()
	return NewSliceSingle(func(ctx context.Context, idx int) (result Merge, err error) {
		resps := f.Call([]reflect.Value{reflect.ValueOf(ctx), keys[idx]})
		if i := resps[0].Interface(); i != nil {
			result = i.(Merge)
		}
		if i := resps[1].Interface(); i != nil {
			err = i.(error)
		}
		return
	}, len(keys))
}
