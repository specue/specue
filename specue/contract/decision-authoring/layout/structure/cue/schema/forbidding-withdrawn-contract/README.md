<dec-body>
На выведенное решение нельзя опираться. Это достигается тем, что у выведенного
решения нет контракта.

```cue
#Decision: {
	status: #DecisionStatus | *"accepted"

	// ...
	if status == "accepted" {
		// действующее — контракт есть
		contract: _            
	}

	if status != "accepted" {
		// выведенное — контракт пуст, 
		// полей нет
		contract: close({})    
	}
}
```

Автору ничего захватывать не нужно — запрет работает сам.
</dec-body>
