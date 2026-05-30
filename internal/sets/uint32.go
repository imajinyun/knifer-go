package sets

import (
	"encoding/json"
	"fmt"
)

type Uint32 map[uint32]struct{}

func NewUint32(items ...uint32) Uint32 {
	sets := make(Uint32, len(items))
	sets.Add(items...)
	return sets
}

func (s Uint32) Add(items ...uint32) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Uint32) Remove(items ...uint32) {
	for _, item := range items {
		_, ok := s[item]
		if ok {
			delete(s, item)
		}
	}
}

func (s Uint32) Contains(item uint32) bool {
	_, ok := s[item]
	return ok
}

func (s Uint32) Sub(ss Uint32) Uint32 {
	out := Uint32{}
	for item := range s {
		if !ss.Contains(item) {
			out[item] = struct{}{}
		}
	}
	return out
}

func (s Uint32) Union(ss Uint32) Uint32 {
	out := Uint32{}
	for item := range s {
		out[item] = struct{}{}
	}
	for item := range ss {
		out[item] = struct{}{}
	}
	return out
}

func (s Uint32) Intersect(ss Uint32) Uint32 {
	return s.Sub(s.Sub(ss))
}

func (s Uint32) Members() []uint32 {
	items := make([]uint32, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

func (s Uint32) Equal(ss Uint32) bool {
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

func (s Uint32) String() string {
	return fmt.Sprintf("set%v", s.Members())
}

func (s Uint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

func (s *Uint32) UnmarshalJSON(data []byte) error {
	var list []uint32
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = NewUint32(list...)
	return nil
}

func (s Uint32) MarshalYAML() (any, error) {
	return s.Members(), nil
}

func (s *Uint32) UnmarshalYAML(unmarshal func(any) error) error {
	var list []uint32
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = NewUint32(list...)
	return nil
}
