<dec-body>

Объявить зависимость в `cue.mod/module.cue`:

```cue
deps: {
    "specue.io/models/decisioning@v0": {v: "v0.0.6"}
}
```

Импортировать пакет с решением в `spec.cue` (`<модуль>/<путь>@<мажор>:<пакет>`):

```cue
import dd "specue.io/models/decisioning/foundation/defining-decision@v0:definingdecision"
```

Решение доступно теперь под алиасом `dd`. 
</dec-body>
