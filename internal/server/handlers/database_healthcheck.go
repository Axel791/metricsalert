package handlers

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type DatabaseHealthCheckHandler struct {
	databaseDSN string
}

func NewDatabaseHealthCheckHandler(databaseDSN string) *DatabaseHealthCheckHandler {
	return &DatabaseHealthCheckHandler{databaseDSN: databaseDSN}
}

func (dh *DatabaseHealthCheckHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	db, err := sqlx.Connect("postgres", dh.databaseDSN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	w.WriteHeader(http.StatusOK)
}
