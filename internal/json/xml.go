package json

import (
	"sort"

	xmlimpl "github.com/imajinyun/knifer-go/internal/xml"
)

// XMLToJSON parses XML text into an ordered JSON object.
func XMLToJSON(xmlStr string) (*JSONObject, error) {
	return XMLToJSONWithOptions(xmlStr)
}

// XMLToJSONWithOptions parses XML text into an ordered JSON object with XML parser options.
func XMLToJSONWithOptions(xmlStr string, opts ...xmlimpl.ParseOption) (*JSONObject, error) {
	m, err := xmlimpl.XMLToMapWithOptions(xmlStr, opts...)
	if err != nil {
		return nil, err
	}
	return mapToJSONObject(m), nil
}

// JSONToXML serializes JSON-compatible data to XML text.
func JSONToXML(root any, rootTag string) (string, error) {
	return JSONToXMLWithOptions(root, rootTag)
}

// JSONToXMLWithOptions serializes JSON-compatible data to XML text with XML writer options.
func JSONToXMLWithOptions(root any, rootTag string, opts ...xmlimpl.WriteOption) (string, error) {
	writeOpts := append([]xmlimpl.WriteOption{xmlimpl.WithRootName(rootTag), xmlimpl.WithOmitDeclaration(true)}, opts...)
	return xmlimpl.MarshalMap(
		jsonValueToMap(root),
		writeOpts...,
	)
}

func mapToJSONObject(m map[string]any) *JSONObject {
	obj := NewJSONObject()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		obj.Set(k, mapXMLValue(m[k]))
	}
	return obj
}

func mapXMLValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return mapToJSONObject(x)
	case []any:
		arr := NewJSONArray()
		for _, item := range x {
			arr.Add(mapXMLValue(item))
		}
		return arr
	case nil:
		return Null
	default:
		return x
	}
}

func jsonValueToMap(root any) map[string]any {
	switch x := root.(type) {
	case *JSONObject:
		return jsonObjectToMap(x)
	case map[string]any:
		return x
	default:
		return map[string]any{"element": jsonValueToPlain(root)}
	}
}

func jsonObjectToMap(obj *JSONObject) map[string]any {
	if obj == nil {
		return nil
	}
	out := make(map[string]any, obj.Len())
	obj.ForEach(func(key string, value any) bool {
		out[key] = jsonValueToPlain(value)
		return true
	})
	return out
}

func jsonValueToPlain(v any) any {
	switch x := v.(type) {
	case *JSONObject:
		return jsonObjectToMap(x)
	case *JSONArray:
		items := x.ToSlice()
		out := make([]any, len(items))
		for i, item := range items {
			out[i] = jsonValueToPlain(item)
		}
		return out
	case jsonNull:
		return nil
	default:
		return v
	}
}
