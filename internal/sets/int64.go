package sets

import (
	"encoding/json"
	"fmt"
)

type Int64 map[int64]struct{}

func NewInt64(items ...int64) Int64 {
	sets := make(Int64, len(items))
	sets.Add(items...)
	return sets
}

func (s Int64) Add(items ...int64) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Int64) Remove(items ...int64) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s Int64) Contains(item int64) bool {
	_, ok := s[item]
	return ok
}

func (s Int64) Sub(ss Int64) Int64 {
	out := Int64{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s Int64) Union(ss Int64) Int64 {
	out := Int64{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s Int64) Intersect(ss Int64) Int64 {
	return s.Sub(s.Sub(ss))
}

func (s Int64) Members() []int64 {
	items := make([]int64, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Int64) Equal(ss Int64) bool {
	for item := range s {
		if !ss.Contains(item) {
			return false
		}
	}
	for item := range ss {
		if !s.Contains(item) {
			return false
		}
	}
	return true
}

func (s Int64) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *Int64) UnmarshalJSON(data []byte) error {
	var list []int64
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewInt64(list...)
	return nil
}

func (s Int64) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *Int64) UnmarshalYAML(unmarshal func(any) error) error {
	var list []int64
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewInt64(list...)
	return nil
}
