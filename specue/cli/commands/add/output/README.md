<dec-body>

Вывод `specue add <path>`. По умолчанию человекочитаемый в `stdout`,
ошибки — в `stderr`. Машинный режим — `--format json`, контракт ниже
держится между реализациями.

## Машинный контракт (`--format json`)

Объект одной из двух форм.

Успех — в каком модуле создано какое решение:
- `ok: true`
- `module` — объект:
  - `id` — идентификатор модуля (`module@version`)
  - `root` — абсолютный путь корня модуля на диске
- `decision` — объект:
  - `dir` — папка решения, путь от корня модуля
  - `structureFile` — созданный файл структуры (`spec.cue`)
  - `bodyFile` — созданный файл тела (`README.md`)

Отказ:
- `ok: false`
- `error` — объект:
  - `code` — `not_a_module` | `invalid_module` | `schema_missing` | `dir_not_empty` | `unknown`
  - `message` — человекочитаемая причина

Коды отказа (проверяются в этом порядке):
- `not_a_module` — по пути нет CUE-модуля (нет `cue.mod/` вверх по дереву);
- `invalid_module` — модуль есть, но CUE не смог его загрузить/разобрать
  (битый `module.cue`, ошибки сборки);
- `schema_missing` — модуль валиден, но не зависит от `specue.io/schema`
  (нечем описывать решения);
- `dir_not_empty` — целевая папка занята;
- `unknown` — дефолтный код для ошибки без своего (usage, неожиданный
  сбой) — по своду cli-guideline.

## Примеры

Успех:

```
$ specue add contract/cli/add
created decision contract/cli/add in /home/user/specue
  contract/cli/add/spec.cue
  contract/cli/add/README.md
```

```
$ specue add contract/cli/add --format json
{"ok": true, "module": {"id": "specue.io/specue@v0", "root": "/home/user/specue"}, "decision": {"dir": "contract/cli/add", "structureFile": "contract/cli/add/spec.cue", "bodyFile": "contract/cli/add/README.md"}}
```

Путь вне модуля решений (`not_a_module`):

```
$ specue add foo/bar --format json
{"ok": false, "error": {"code": "not_a_module", "message": "./foo/bar is not a specue decision module"}}
```

Модуль есть, но невалиден (`invalid_module`):

```
$ specue add contract/cli/add --format json
{"ok": false, "error": {"code": "invalid_module", "message": "module at /home/user/specue failed to load: ..."}}
```

Нет зависимости от схемы (`schema_missing`):

```
$ specue add contract/cli/add --format json
{"ok": false, "error": {"code": "schema_missing", "message": "module does not depend on specue.io/schema, used to describe decisions"}}
```

Папка не пуста (`dir_not_empty`):

```
$ specue add contract/cli/add --format json
{"ok": false, "error": {"code": "dir_not_empty", "message": "./contract/cli/add is not empty"}}
```
</dec-body>
