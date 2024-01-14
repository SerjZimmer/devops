package storage

import (
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type MetricsStorage struct {
	*MetricsStorageInternal
}

func (s *MetricsStorage) UpdateMetricValue(m Metrics) error {
	return Retry(func() error {
		return s.MetricsStorageInternal.UpdateMetricValue(m)
	})
}

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
