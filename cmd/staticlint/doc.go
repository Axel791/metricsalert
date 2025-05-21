/*
Package staticlint предоставляет одноимённый multichecker.

# Запуск

	# проанализировать текущий проект
	go run ./cmd/staticlint ./...

	# запустить только проверку SA1000
	staticlint -SA1000 ./...

# Состав анализаторов

 1. **golang.org/x/tools/go/analysis/passes** – стандартные проверки Go.
 2. **SA**-семейство Staticcheck – обнаружение потенциальных багов.
 3. **ST1019** – валидность HTTP-статусов (пример другого класса Staticcheck).
 4. **errwrap** – контроль передачи %w в fmt.Errorf.
 5. **goone** – запрет тяжёлых SQL-запросов внутри циклов.
 6. **exitinmain** – кастомный анализатор, запрещающий прямой os.Exit
    в функции main (см. пакет exitinmain).

Каждый анализатор можно включать/выключать ключами `-<NAME>`, подробности —
`staticlint -help`.
*/
package main
