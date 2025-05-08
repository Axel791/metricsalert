package handlers

import (
	"html/template"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"
)

// GetMetricsHTMLHandler renders the current set of metrics as a simple HTML page.
//
// The handler responds to **GET /** (или другой настроенный путь) и возвращает
// страницу‑таблицу, пригодную для быстрого визуального просмотра значений.
//
// # Таблица полей
// | Столбец | Смысл                                   |
// |---------|-----------------------------------------|
// | Имя     | ID метрики                              |
// | Тип     | Go‑тип отображаемого значения (int64/float64) |
// | Значение| Числовое значение метрики               |
//
// # Пример запроса
//
//	GET / HTTP/1.1
//
// # Пример ответа (фрагмент)
//
//	<tr><td>Alloc</td><td>float64</td><td>6.27</td></tr>
//
// # Возможные коды ошибок
//
//	500 – при неуспехе services.Metric.GetAllMetric или ошибке шаблона.
//
// GetMetricsHTMLHandler **не кэширует** результат и каждый раз берёт свежие данные
// из переданной реализации services.Metric.
//
// Экземпляр безопасен для конкурентного использования.
type GetMetricsHTMLHandler struct {
	metricService services.Metric
}

// NewGetMetricsHTMLHandler возвращает инициализированный HTML‑хендлер.
func NewGetMetricsHTMLHandler(metricService services.Metric) *GetMetricsHTMLHandler {
	return &GetMetricsHTMLHandler{metricService: metricService}
}

// ServeHTTP реализует интерфейс http.Handler.
//
// Алгоритм:
//  1. Получает все метрики через MetricService.
//  2. Преобразует их в map[string]interface{} для удобной передачи в шаблон.
//  3. Парсит встроенный html/template и рендерит его в ответ.
func (h *GetMetricsHTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricService.GetAllMetric(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metricsMap := make(map[string]interface{})
	for _, metric := range metrics {
		var value interface{}
		switch metric.MType {
		case domain.Counter:
			value = metric.Delta.Int64
		case domain.Gauge:
			value = metric.Value.Float64
		default:
			value = "unknown"
		}
		metricsMap[metric.ID] = value
	}

	const tpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Метрики</title>
</head>
<body>
    <h1>Список метрик</h1>
    <table border="1">
        <tr>
            <th>Имя:</th>
            <th>Тип:</th>
            <th>Значение:</th>
        </tr>
        {{ range $name, $value := . }}
        <tr>
            <td>{{ $name }}</td>
            <td>{{ printf "%T" $value }}</td>
            <td>{{ $value }}</td>
        </tr>
        {{ end }}
    </table>
</body>
</html>`

	tmpl, err := template.New("metrics").Parse(tpl)
	if err != nil {
		http.Error(w, "Ошибка при генерации страницы", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, metricsMap); err != nil {
		http.Error(w, "Ошибка при генерации страницы", http.StatusInternalServerError)
		return
	}
}
