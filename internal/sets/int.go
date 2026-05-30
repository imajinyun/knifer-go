package sets

import (
	"encoding/json"
	"fmt"
)

type Int map[int]struct{}

func NewInt(items ...int) Int {
	sets := make(Int, len(items))
	sets.Add(items...)
	return sets
}

func (s Int) Add(items ...int) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Int) Remove(items ...int) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s Int) Contains(item int) bool {
	_, ok := s[item]
	return ok
}

func (s Int) Sub(ss Int) Int {
	out := Int{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s Int) Union(ss Int) Int {
	out := Int{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s Int) Intersect(ss Int) Int {
	return s.Sub(s.Sub(ss))
}

func (s Int) Members() []int {
	items := make([]int, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Int) Equal(ss Int) bool {
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

func (s Int) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s Int) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *Int) UnmarshalJSON(data []byte) error {
	var list []int
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewInt(list...)
	return nil
}

func (s Int) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *Int) UnmarshalYAML(unmarshal func(any) error) error {
	var list []int
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewInt(list...)
	return nil
}
