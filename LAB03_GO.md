# Лабораторная работа 3 Go

Обязательные части задания адаптированы под текущий backend на Go.

1. Жизненный цикл тестов: `backend/internal/domain/users/main_test.go` использует `TestMain` — аналог suite-level setup/teardown. Текущие unit-тесты используют memory store и не требуют внешнего драйвера.
2. Параметризация: `TestRegisterRejectsInvalidInput` использует таблицу кейсов и `t.Run`, что является Go-аналогом Data Provider.
3. Группы: smoke запускается `make test-smoke`; полный регрессионный набор — `make test-regression`.
4. CI/CD: `.github/workflows/ci.yml` запускает smoke на каждом push, regression на pull request в `main` и ночью по расписанию. HTML-отчёт покрытия и `coverage.out` сохраняются как artifacts.
5. Git: для сдачи создай ветку `feature/lab03-tests`, закоммить эти изменения, выполни push и открой Pull Request в `main`. Pull Request нельзя создать локально без авторизации и отправки ветки на GitHub.

Бонусные пункты ThreadLocal и Retry намеренно не добавлены: это Java/Python-специфика и они не нужны для обязательной Go-версии.
