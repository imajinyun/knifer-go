package sets

import (
	"encoding/json"
	"fmt"
)

type Uint map[uint]struct{}

func NewUint(items ...uint) Uint {
	sets := make(Uint, len(items))
	sets.Add(items...)
	return sets
}

func (s Uint) Add(items ...uint) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Uint) Remove(items ...uint) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s Uint) Contains(item uint) bool {
	_, ok := s[item]
	return ok
}

func (s Uint) Sub(ss Uint) Uint {
	out := Uint{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s Uint) Union(ss Uint) Uint {
	out := Uint{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s Uint) Intersect(ss Uint) Uint {
	return s.Sub(s.Sub(ss))
}

func (s Uint) Members() []uint {
	items := make([]uint, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Uint) Equal(ss Uint) bool {
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

func (s Uint) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s Uint) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *Uint) UnmarshalJSON(data []byte) error {
	var list []uint
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewUint(list...)
	return nil
}

func (s Uint) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *Uint) UnmarshalYAML(unmarshal func(any) error) error {
	var list []uint
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewUint(list...)
	return nil
}
