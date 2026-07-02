# First-Use Golden Paths

Use this page for the first ten minutes with `knifer-go`: 10 tasks in 10 minutes.
Each task gives one
recommended facade and one shortest example. Follow the single facade first;
move to related packages only after the first path works.

## 10 Tasks In 10 Minutes

| Task | Recommended facade | Shortest example |
| --- | --- | --- |
| string | `vstr` | `vstr.DefaultIfBlank(" ", "guest")` |
| slice | `vslice` | `vslice.Map([]int{1, 2}, func(n int) int { return n * 2 })` |
| map | `vmap` | `vmap.Pick(map[string]int{"a": 1, "b": 2}, "a")` |
| json | `vjson` | `vjson.GetByPath(obj, "user.name")` |
| file | `vfile` | `vfile.ReadFileString("config.txt")` |
| http | `vhttp` | `vhttp.Get("https://example.com").Execute()` |
| crypto | `vcrypto` | `vcrypto.SHA256Hex("hello")` |
| config | `vconf` | `vconf.Parse("app.port=8080\\n")` |
| db | `vdb` | `vdb.Select("id").From("users").Where(vdb.Eq("active", true)).SQL()` |
| cli | `vcli` | `vcli.NewFlagParser("serve").Parse([]string{"--help"})` |

## Examples

### string

```go
fmt.Println(vstr.DefaultIfBlank(" ", "guest"))
```

### slice

```go
fmt.Println(vslice.Map([]int{1, 2}, func(n int) int { return n * 2 }))
```

### map

```go
fmt.Println(vmap.Pick(map[string]int{"a": 1, "b": 2}, "a"))
```

### json

```go
obj, _ := vjson.ParseObj(`{"user":{"name":"knifer-go"}}`)
fmt.Println(vjson.GetByPath(obj, "user.name"))
```

### file

```go
text, err := vfile.ReadFileString("config.txt")
fmt.Println(text, err)
```

### http

```go
resp := vhttp.Get("https://example.com").Execute()
defer resp.Close()
fmt.Println(resp.Status(), resp.Err())
```

### crypto

```go
fmt.Println(vcrypto.SHA256Hex("hello"))
```

### config

```go
cfg, _ := vconf.Parse("app.port=8080\n")
fmt.Println(cfg.GetInt("app.port", 0))
```

### db

```go
sql, args, _ := vdb.Select("id").From("users").Where(vdb.Eq("active", true)).SQL()
fmt.Println(sql, args)
```

### cli

```go
parser := vcli.NewFlagParser("serve")
port := parser.Int("port", 8080, "port to bind")
_, _ = parser.Parse([]string{"--port", "9090"})
fmt.Println(*port)
```

## Rules

- Start with one recommended facade per task.
- Keep examples short enough to paste into a scratch file.
- Use explicit error-returning flows before defaults.
- Use Safe context-aware or WithOptions flows for trust boundaries when the
  input crosses a trust boundary.
