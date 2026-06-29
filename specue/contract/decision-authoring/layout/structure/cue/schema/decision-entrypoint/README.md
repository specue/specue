<dec-body>

Решение в CUE-файле — это поле верхнего уровня `decision` типа `#Decision`:

```cue
package whatever

decision: s.#Decision & {
	problem: "..."
	// ...
}
```

Имя поля `decision` фиксировано — по нему загрузчик находит решение
в пакете. Имя пакета произвольно и идентификатором не является.
</dec-body>
