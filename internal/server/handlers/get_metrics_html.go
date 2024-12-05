package handlers

import (
	"html/template"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/storage"
)

type GetMetricsHTMLHandler struct {
	storage storage.Store
}

func NewGetMetricsHTMLHandler(storage storage.Store) *GetMetricsHTMLHandler {
	return &GetMetricsHTMLHandler{storage: storage}
}

func (h *GetMetricsHTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := h.storage.GetAllMetrics()

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
	if err := tmpl.Execute(w, metrics); err != nil {
		http.Error(w, "Ошибка при генерации страницы", http.StatusInternalServerError)
		return
	}
}
