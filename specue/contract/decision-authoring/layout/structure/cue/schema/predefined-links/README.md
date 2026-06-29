<dec-body>

Известные виды связей задаются предопределёнными обёртками над `#Link`
с зафиксированным `kind`:

```cue
#Supersedes: #Link & {kind: "supersedes"}
```

Автор пишет `#Supersedes & {to: ...}` вместо `#Link & {kind: "supersedes", to: ...}`.
</dec-body>
