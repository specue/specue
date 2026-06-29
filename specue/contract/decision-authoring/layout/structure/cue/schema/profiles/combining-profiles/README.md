<dec-body>

Несколько профилей совмещаются конъюнкцией в `context` — каждый описывает свою часть:

```cue
decision: s.#Decision & {
	problem: "..."
	context: sh.#WithDrivers & dm.#WithDimensions & {
		drivers: [...]
		dimensions: {system: layer: "domain"}
	}
}
```
</dec-body>
