package api

// Metrics описывает универсальный JSON‑контейнер для передачи значения одной
// метрики (gauge или counter) между клиентом и сервером.
//
// # Поля
//   - ID     — уникальное строковое имя метрики.
//   - MType  — тип метрики: "gauge" либо "counter".
//   - Delta  — при MType=="counter" содержит новое приращение (Int64). Отсутствует
//     или null в запросах/ответах gauge‑метрик.
//   - Value  — при MType=="gauge" содержит само числовое значение (Float64).
//     Отсутствует или null в запросах/ответах counter‑метрик.
//
// В каждом экземпляре одновременно задано **только одно** из полей Delta/Value.
// Сервер обязан валидировать согласованность: переданное поле должно
// соответствовать объявленному MType, иначе возвращается HTTP 400.
//
// # Пример (gauge)
//
//	{
//	  "id":    "Alloc",
//	  "type":  "gauge",
//	  "value": 6.27
//	}
//
// # Пример (counter)
//
//	{
//	  "id":    "PollCount",
//	  "type":  "counter",
//	  "delta": 3
//	}
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// GetMetric описывает запрос клиента на получение значения конкретной метрики.
// Используется хендлером GetMetricHandler.
//
// Обязательные поля:
//   - ID    — уникальное имя метрики;
//   - MType — ожидаемый тип ("gauge" или "counter").
//
// Пример:
//
//	{
//	  "id":   "Alloc",
//	  "type": "gauge"
//	}
type GetMetric struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}
