# 🐹 Go Backend: практический курс для junior-разработчика

13 мини-проектов с mock-данными. В каждом уже есть код с **багами** 🐞 и **заготовками** ✍️: их нужно исправить или дописать, чтобы сервис отдавал правильные данные. Проверяет работу **симулятор CI-пайплайна**, похожий на GitHub Actions: `gofmt → build → vet → tests (-race) → coverage`.

У каждого модуля свой `README.md` с теорией, таблицей заданий и ожидаемыми данными, подсказками и **решениями под спойлером** 🔒.

## 🗺 Программа курса

| # | Модуль | Темы | Формат |
|---|---|---|---|
| **Часть 1. Язык** ||||
| 01 | [structs-slices](01-structs-slices/) | структуры, слайсы, указатели, копия vs оригинал, nil vs пустой слайс, пагинация | 3 🐞 + 2 ✍️ |
| 02 | [maps-aggregation](02-maps-aggregation/) | map, nil-map, «comma ok», set, сортировка, топ-N | 3 🐞 + 2 ✍️ |
| 03 | [errors](03-errors/) | sentinel-ошибки, `%w`, `errors.Is/As`, свои типы ошибок, ошибка → HTTP-статус | 2 🐞 + 2 ✍️ |
| 04 | [interfaces-repository](04-interfaces-repository/) | интерфейсы, value/pointer receiver, Repository + Service, Stringer | 3 🐞 + 2 ✍️ |
| 05 | [json](05-json/) | теги, `omitempty`, `json:"-"`, DTO, конверт ответа, `DisallowUnknownFields`, `go:embed` | 2 🐞 + 2 ✍️ |
| **Часть 2. HTTP** ||||
| 06 | [http-handlers](06-http-handlers/) | `net/http`, роутинг Go 1.22, path/query-параметры, статус-коды, `httptest` | 3 🐞 + 2 ✍️ |
| 07 | [middleware](07-middleware/) | middleware-цепочки, logger, recover, request-id, auth (401/403), `context.Value` | 3 🐞 + 3 ✍️ |
| 08 | [rest-crud-validation](08-rest-crud-validation/) | CRUD, валидация по полям, UTF-8, пагинация, PATCH через указатели, 204 | 3 🐞 + 2 ✍️ |
| **Часть 3. Конкурентность** ||||
| 09 | [goroutines-mutex](09-goroutines-mutex/) | горутины, WaitGroup, Mutex/RWMutex, `sync.Once`, детектор гонок | 4 🐞 + 1 ✍️ |
| 10 | [channels-worker-pool](10-channels-worker-pool/) | каналы, pipeline, fan-in, worker pool, `select`, таймауты, утечки горутин | 2 🐞 + 3 ✍️ |
| 11 | [context-timeouts](11-context-timeouts/) | `context`: отмена, дедлайны, `r.Context()`, 504, `go vet lostcancel` | 3 🐞 + 1 ✍️ |
| **Часть 4. Инженерия** ||||
| 12 | [testing-mocks](12-testing-mocks/) | DI, stub/fake/spy, table-driven тесты, **порог покрытия 90%** | 3 🐞 + свои тесты |
| 13 | [final-orders-api](13-final-orders-api/) | 🏁 итоговый проект: слои domain/repo/service/http, машина состояний, e2e-тесты, graceful shutdown | 4 🐞 + 1 ✍️ |

Обозначения: 🐞 — найти и исправить баг, ✍️ — дописать код.

## ⚙️ Требования
- **Go 1.22+** (проверено на Go 1.25). Проверить: `go version`.
- Внешних зависимостей нет, только стандартная библиотека.
- `make` по желанию: все команды можно запускать через `go run`.

## 🚦 Как работает пайплайн

Все команды запускаются **из корня курса** (эта папка):

```bash
go run ./tools/ci 01      # пайплайн одного модуля
make ci M=01              # то же самое через make
make ci-all               # все модули + сводка «Пройдено модулей: X/13»
```

Пример вывода:
```
════════════════════════════════════════════════════════
  CI Pipeline · 01-structs-slices
════════════════════════════════════════════════════════
[1/4] 🧹 gofmt     ✅ passed (0.0s)
[2/4] 🔨 build     ✅ passed (0.3s)
[3/4] 🔍 vet       ✅ passed (0.2s)
[4/4] 🧪 tests     ❌ failed (1.1s)
      ✅ TestActiveUsers
      ❌ TestApplyBonus
           │ users_test.go:51: баланс пользователя #1 (Анна) = 1500, ожидалось 1600
      ...
      Итог: 4/6 тестов пройдено
════════════════════════════════════════════════════════
  ❌ PIPELINE FAILED на стадии tests
```

| Стадия | Что делает | Как в реальном CI |
|---|---|---|
| 🧹 gofmt | код отформатирован (`gofmt -l`) | ревьюер не примет неформатированный код |
| 🔨 build | код компилируется | — |
| 🔍 vet | статический анализ `go vet` (+ компиляция тестов) | линтеры (`golangci-lint`) |
| 🧪 tests | `go test -race -count=1 -timeout 60s` | юнит- и интеграционные тесты, детектор гонок |
| 📊 coverage | покрытие ≥ порога из `ci.cfg` (только модуль 12) | quality gate (SonarQube, Codecov) |

Если стадия упала, следующие **пропускаются** (`⏭ skipped`), как в настоящем CI. Модуль выполнен, когда видишь `✅ PIPELINE PASSED`.

> Некоторые модули **специально** падают не на тестах. Модуль 04 падает на `build`, 05 и 11 — на `vet`, 12 — на `vet`, а потом на `coverage`. Это часть заданий: учимся читать ошибки компилятора и линтера.

## 🧭 Как проходить модуль

1. Открой `NN-.../README.md` и прочитай **теорию**.
2. Запусти пайплайн `go run ./tools/ci NN` и посмотри, что падает.
3. Исправляй код в файлах, помеченных в README как «Работаешь здесь». **Тесты (`*_test.go`) не меняй**, единственное исключение — `12-testing-mocks/student_test.go`.
4. Застрял? Открой **💡 Подсказки**. Совсем застрял? Загляни в **🔒 Решение**, но сначала попробуй сам.
5. Гоняй пайплайн, пока не увидишь `✅ PIPELINE PASSED`.
6. Ответь себе на **🧠 вопросы для самопроверки**: их часто задают на собеседованиях.

Полезные команды:
```bash
make run M=06                                    # запустить модуль
make test M=09                                   # подробный вывод тестов
go test -race -run TestPaginate -v ./01-structs-slices   # один тест
make fmt M=01                                    # починить стадию gofmt
```

💾 **Совет:** перед стартом сделай `git init && git add . && git commit -m "start"`. Тогда `git diff` покажет твои изменения, а `git checkout -- 01-structs-slices` вернёт модуль в исходное состояние.

## 📁 Структура
```
.
├── go.mod                 ← один модуль на весь курс (module gocourse)
├── Makefile
├── tools/ci/main.go       ← симулятор CI-пайплайна
├── 01-structs-slices/
│   ├── README.md          ← теория, задания, подсказки, решения
│   ├── users.go           ← код с багами и TODO (здесь работаешь)
│   ├── mock_data.go       ← mock-данные
│   ├── main.go            ← демо-запуск
│   └── users_test.go      ← автотесты пайплайна
├── ...
└── 13-final-orders-api/
    └── internal/{domain,repo,service,httpapi}/
```
