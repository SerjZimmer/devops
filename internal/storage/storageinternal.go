package storage

import (
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// MetricsStorage представляет собой структуру, обертывающую внутреннюю структуру MetricsStorageInternal.
type MetricsStorage struct {
	*MetricsStorageInternal
}

// UpdateMetricValue обновляет значение метрики с использованием механизма повторных попыток.
func (s *MetricsStorage) UpdateMetricValue(m Metrics) error {
	return Retry(func() error {
		return s.MetricsStorageInternal.UpdateMetricValue(m)
	})
}

// Retry выполняет функцию fn с повторными попытками в случае ошибок, связанных с подключением к базе данных.
func Retry(fn func() error) error {
	err := fn()
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		if err != nil {
			try := 1
			for err != nil && try < 4 {
				time.Sleep(time.Duration(2*(try-1) + 1))
				err = fn()
				try++
			}
		}

	}
	return err
}
