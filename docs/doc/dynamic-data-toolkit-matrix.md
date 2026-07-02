# Dynamic Data Toolkit Matrix

Use this page when data shape is not fully known at compile time. The goal is
to choose the narrowest facade that owns the current boundary instead of
placing all dynamic behavior in one broad helper package.

## Matrix

| Workflow | Start with | Specialist comparison | Boundary |
| --- | --- | --- | --- |
| configuration loading | `vconf` | `spf13/viper`, `mapstructure` decode paths | Use when input is config text, files, profiles, environment expansion, remote config, schema validation, or safe remote loading. |
| map/struct decode | `vbean` | `mitchellh/mapstructure` | Use when map or struct data must bind into caller-owned structs with weak conversion, hooks, strict unused reporting, or field-path errors. |
| struct copy | `vbean` | `jinzhu/copier` | Use when trusted Go values need tag-aware copy, merge, or map/struct shape conversion without JSON serialization. |
| JSON object path | `vjson` | `encoding/json`, JSON query helpers | Use when small in-memory JSON documents need object, array, path, formatting, or XML bridge helpers. |
| dynamic object checks | `vobj` | `thoas/go-funk`, reflection-heavy helpers | Use when `any` values need nil/empty checks, length, membership, defaults, comparison, or serialization-based cloning. |
| reflection field access | `vref` | direct `reflect`, `go-funk` reflection helpers | Use when fields, tags, constructors, methods, or call targets are discovered at runtime. |
| scalar conversion after lookup | `vconv` | `spf13/cast` | Use after dynamic lookup when invalid scalar input must be explicit or fallback behavior must be visible. |

## Decision Rules

- Use typed Go code when the shape is known.
- Use `vconf` before `vbean` when data is still configuration input.
- Use `vbean` before `vobj` when tag-aware copy, decode, merge, or unused-key
  metadata matters.
- Use `vjson` for JSON-shaped object/path work; use `encoding/json.Decoder`
  directly for streams, token-level control, or strict decoder policy.
- Use `vref` only at dynamic adapter boundaries; keep normal business logic
  typed.
- Use `vconv` for scalar conversion after dynamic lookup, not for collection or
  struct mapping.
- Avoid reflection-heavy hot paths until typed alternatives and benchmarks have
  been reviewed.

## Cookbook

### configuration loading

```go
cfg, err := vconf.Parse("server.port=8080\n")
if err != nil {
	return err
}
port := cfg.GetInt("server.port", 0)
_ = port
```

### map/struct decode

```go
var dst struct {
	Port int `bean:"port"`
}
err := vbean.Decode(map[string]any{"port": "8080"}, &dst)
_ = err
```

### JSON object path

```go
obj, err := vjson.ParseObj(`{"user":{"name":"knifer-go"}}`)
if err != nil {
	return err
}
name := vjson.GetByPath(obj, "user.name")
_ = name
```

### dynamic object checks

```go
empty := vobj.IsEmpty([]string{})
length := vobj.Length(map[string]int{"go": 1})
_ = empty
_ = length
```

### reflection field access

```go
value := vref.GetFieldValue(struct{ Name string }{Name: "alice"}, "Name")
_ = value
```

### scalar conversion after lookup

```go
port, err := vconv.ToIntE("8080")
_ = port
_ = err
```

## Machine-Readable Boundaries

- configuration loading
- map/struct decode
- struct copy
- JSON object path
- dynamic object checks
- reflection field access
- scalar conversion after lookup
- typed Go code first
- vconf before vbean for configuration input
- vbean before vobj for mapping metadata
- vref only at dynamic adapter boundaries
- vconv after dynamic lookup
- avoid reflection-heavy hot paths
