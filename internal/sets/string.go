package sets

import (
	"encoding/json"
	"fmt"
)

type String map[string]struct{}

func NewString(items ...string) String {
	sets := make(String, len(items))
	sets.Add(items...)
	return sets
}

func (s String) Add(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s String) Remove(items ...string) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s String) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s String) Sub(ss String) String {
	out := String{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s String) Union(ss String) String {
	out := String{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s String) Intersect(ss String) String {
	return s.Sub(s.Sub(ss))
}

func (s String) Members() []string {
	items := make([]string, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s String) Equal(ss String) bool {
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

func (s String) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *String) UnmarshalJSON(data []byte) error {
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewString(list...)
	return nil
}

func (s String) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *String) UnmarshalYAML(unmarshal func(any) error) error {
	var list []string
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewString(list...)
	return nil
}
