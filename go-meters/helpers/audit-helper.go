package helpers

import (
	"database/sql"
	"sync"
)

type AuditLog struct {
	sync.RWMutex
	db *sql.DB
}

func (a *AuditLog) SetDb(db *sql.DB) {
	a.Lock()
	defer a.Unlock()

	a.db = db
}

func (a AuditLog) Write(p []byte) (n int, err error) {
	query := `
        INSERT INTO wmeter.audit_log (
            audit_msg
        )
        VALUES (
            $1
        )
    `

	msg := string(p)

	_, err = a.db.Exec(query, msg)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}
