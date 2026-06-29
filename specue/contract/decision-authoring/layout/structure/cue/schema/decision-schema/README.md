<dec-body>

Решение описывается типом `#Decision`:

```cue
#Decision: {
	// идентичность решения
	problem:  string             

	status:   #DecisionStatus | *"accepted" 

	links: [...#Link]            

	// открытый контекст для профилей
	context: {...}               

	// Контракт только у действующего решения.
	// Слот открыт (`_`)  
	// Автор должен явно его закрывать 
	// сам с помощью close()
	// `contract: close({...})`. 
	if status == "accepted" {
		contract: _
	}

	if status != "accepted" {
		// у выведенного решения контракта нет
		contract: close({})      
	}
}
```

`#Decision` встраивает форму контракта, статуса, связи и контекст — детали
каждого в своих узлах.

У выведенного решения (`superseded`/`retired`) контракта нет.
</dec-body>
