package sets

import (
	"encoding/json"
	"fmt"
)

type Int32 map[int32]struct{}

func NewInt32(items ...int32) Int32 {
	sets := make(Int32, len(items))
	sets.Add(items...)
	return sets
}

func (s Int32) Add(items ...int32) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Int32) Remove(items ...int32) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s Int32) Contains(item int32) bool {
	_, ok := s[item]
	return ok
}

func (s Int32) Sub(ss Int32) Int32 {
	out := Int32{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s Int32) Union(ss Int32) Int32 {
	out := Int32{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s Int32) Intersect(ss Int32) Int32 {
	return s.Sub(s.Sub(ss))
}

func (s Int32) Members() []int32 {
	items := make([]int32, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Int32) Equal(ss Int32) bool {
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

func (s Int32) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s Int32) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *Int32) UnmarshalJSON(data []byte) error {
	var list []int32
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewInt32(list...)
	return nil
}

func (s Int32) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *Int32) UnmarshalYAML(unmarshal func(any) error) error {
	var list []int32
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewInt32(list...)
	return nil
}
