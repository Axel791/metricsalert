package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgconn"
)

// RetryOperation выполняет операцию с повторными попытками при ошибке транспорта.
func RetryOperation(operation func() error) error {
	retries := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i, delay := range retries {
		err := operation()
		if err == nil {
			return nil
		}

		if isConnectionException(err) {
			if i < len(retries)-1 {
				time.Sleep(delay)
				continue
			}
		}

		return err
	}

	return fmt.Errorf("operation failed after retries")
}

func isConnectionException(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08")
	}
	return false
}
