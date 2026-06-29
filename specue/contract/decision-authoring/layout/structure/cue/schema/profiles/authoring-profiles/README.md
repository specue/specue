<dec-body>

Профиль — пара CUE-определений в своём пакете: тип данных и
подмешиваемое `#WithX`, описывающее форму этих данных в `context`.

```cue
#Dimensions: {[string]: {[string]: string}, ...}

// профиль описывает рашсирение для context
#WithDimensions: {
	dimensions: #Dimensions
	...
}
```

Автор применяет профиль к `context`: 
`context: #WithDimensions & {...}`. 

</dec-body>
