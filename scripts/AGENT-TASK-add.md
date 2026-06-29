# Задание: привести реализацию `specue add` в соответствие с решениями

Источник истины — граф решений. Полный свод достижимых из вершины `binary`
решений собран в `DECISIONS.md` (37 узлов: README = поведение, spec.cue =
контракт). Реализация корректна, когда проходит conformance-набор
(`mise run conformance`).

## Что делать

1. Прочитать `DECISIONS.md` — это контракт, по которому работает `add`.
2. Прогнать conformance (`conformance/testdata/*.txtar`) — они УЖЕ обновлены
   под актуальный контракт и сейчас расходятся с кодом.
3. Привести `cmd/`, `add.go` и связанный код так, чтобы набор прошёл.
4. НИЧЕГО не менять в txtar/решениях — они истина; меняется только реализация.

## Что изменилось в контракте (сверь с прежней реализацией)

### Машинный вывод `add --format json` — НОВАЯ структура
Было (плоско):
```json
{"ok":true,"id":"...","module":"...@v0","root":"...","created":["...","..."]}
```
Стало (сгруппировано — где модуль / что за решение):
```json
{
  "ok": true,
  "module":   {"id": "specue.io/demo@v0", "root": "<abs-path>"},
  "decision": {"dir": "domain/foo",
               "structureFile": "domain/foo/spec.cue",
               "bodyFile": "domain/foo/README.md"}
}
```
- `id`/`module`/`root`/`created` БОЛЬШЕ НЕТ на верхнем уровне;
- `module.id` = module@version, `module.root` = абсолютный путь корня;
- `decision.dir` = путь решения от корня модуля (бывш. `id`);
- созданные файлы — два ИМЕНОВАННЫХ поля `structureFile`/`bodyFile`
  (а не массив `created`).

Форма отказа НЕ изменилась:
```json
{"ok": false, "error": {"code": "<код>", "message": "<человекочитаемо>"}}
```

### Код отказа «папка занята»
`dir_occupied` → **`dir_not_empty`** (канон — как в тестах). Других кодов
не трогали: `not_a_module`, `invalid_module`, `schema_missing`.

### Ответственность шагов (если код это отражает в структуре)
- `resolving-module` теперь ТОЛЬКО находит корень модуля (отказ
  `not_a_module`);
- проверка целостности (`invalid_module`, `schema_missing`) — при
  ЗАГРУЗКЕ модуля (`loading-module`), от найденного корня;
- проверка «папка занята» (`dir_not_empty`) — при создании скелета
  (`creating-decision-package`).

### Команда не падает молча (НОВЫЙ сценарий `add_missing_arg`)
`specue add` без обязательного `<path>` сейчас выходит с кодом 1, но
НИЧЕГО не печатает (cobra `SilenceErrors: true` в `cmd/root.go` глушит
ошибку, а `Execute()` её не выводит). По своду это запрещено:
- отсутствие обязательного аргумента — ошибка с сообщением в stderr
  (не молчание);
- в `--format json` — та же форма `{ok:false, error:{code, message}}`,
  код по умолчанию `unknown` (у usage-ошибки нет своего доменного кода);
- `unknown` — общий дефолтный код для ошибок без своего (usage,
  неожиданный сбой).
Почини вывод ошибок: usage/неизвестные ошибки должны печататься
(human → stderr; json → errorShape с `code: "unknown"`), а не глохнуть.

## Проверка
`mise run conformance` зелёный = реализация соответствует решениям.
Точные jq-ассерты формы — в `conformance/testdata/add_success.txtar` и
`add_*.txtar` (включая `add_missing_arg.txtar`).
