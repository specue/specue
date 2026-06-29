<dec-body>

CUE-пакет загружается через Go API CUE:

```go
insts := load.Instances([]string{"."}, &load.Config{Dir: dir})
v := ctx.BuildInstance(insts[0])
```

`load.Instances` находит пакет по пути, `BuildInstance` вычисляет его в
`cue.Value`. 
</dec-body>
