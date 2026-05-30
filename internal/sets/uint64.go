package sets

import (
	"encoding/json"
	"fmt"
)

type Uint64 map[uint64]struct{}

func NewUint64(items ...uint64) Uint64 {
	sets := make(Uint64, len(items))
	sets.Add(items...)
	return sets
}

func (s Uint64) Add(items ...uint64) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Uint64) Remove(items ...uint64) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s Uint64) Contains(item uint64) bool {
	_, ok := s[item]
	return ok
}

func (s Uint64) Sub(ss Uint64) Uint64 {
	out := Uint64{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s Uint64) Union(ss Uint64) Uint64 {
	out := Uint64{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s Uint64) Intersect(ss Uint64) Uint64 {
	return s.Sub(s.Sub(ss))
}

func (s Uint64) Members() []uint64 {
	items := make([]uint64, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Uint64) Equal(ss Uint64) bool {
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

func (s Uint64) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *Uint64) UnmarshalJSON(data []byte) error {
	var list []uint64
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewUint64(list...)
	return nil
}

func (s Uint64) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *Uint64) UnmarshalYAML(unmarshal func(any) error) error {
	var list []uint64
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewUint64(list...)
	return nil
}
