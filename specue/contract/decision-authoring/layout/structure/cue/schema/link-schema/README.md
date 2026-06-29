<dec-body>
Связь описывается типом `#Link`:

```cue
#Link: {
	kind: "supersedes" | string
	to:   #Decision
}
```

`kind` — вид связи (открыт строкой), `to` — целевое решение.
</dec-body>
