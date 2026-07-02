package storage

import (
	"database/sql"
	"fmt"
)

func TimestampPtr(value sql.NullInt64) *Timestamp {
	if !value.Valid {
		return nil
	}
	ts := Timestamp(value.Int64)
	return &ts
}

func NullableTimestamp(value *Timestamp) any {
	if value == nil {
		return nil
	}
	return *value
}

func RequireAffected(result sql.Result, notFound error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}
	if affected == 0 {
		return notFound
	}
	return nil
}
