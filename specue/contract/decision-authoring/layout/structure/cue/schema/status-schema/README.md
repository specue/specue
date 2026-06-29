<dec-body>

Статус решения описан типом `#DecisionStatus` — всё множество статусов:

```cue
// действующие
#Active:    "accepted"

// выведенные
#Withdrawn: "superseded" | "retired" 

// всё множество — объединение подмножеств
#DecisionStatus: #Active | #Withdrawn
```

Статусы заданы двумя подмножествами: `#Active` (решение действует) и
`#Withdrawn` (решение выведено).

По умолчанию статус решения — `accepted`.
</dec-body>
