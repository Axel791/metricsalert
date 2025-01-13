package handlers

import (
	"html/template"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/services"
)

type GetMetricsHTMLHandler struct {
	metricService services.Metric
}

func NewGetMetricsHTMLHandler(metricService services.Metric) *GetMetricsHTMLHandler {
	return &GetMetricsHTMLHandler{metricService: metricService}
}

func (h *GetMetricsHTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricService.GetAllMetric(r.Context())

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metricsMap := make(map[string]interface{})
	for _, metric := range metrics {
		var value interface{}
		if metric.MType == domain.Counter {
			value = metric.Delta.Int64
		} else if metric.MType == domain.Gauge {
			value = metric.Value.Float64
		} else {
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
    </html>
    `
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
