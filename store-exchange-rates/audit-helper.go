package main

import "sync"
import "time"

type AuditLog struct {
	sync.RWMutex
}

func (a AuditLog) Write(p []byte) (n int, err error) {
	query := prepareQuery(`
        INSERT INTO audit_log (
            log_time, audit_msg
        )
        VALUES (?, ?)
    `)

	logTime := time.Now().UTC()
	msg := string(p)

	_, err = db.Exec(query, logTime, msg)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}
