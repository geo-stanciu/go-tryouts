package main

import "sync"

type AuditLog struct {
	sync.RWMutex
}

func (a AuditLog) Write(p []byte) (n int, err error) {
	query := `
        INSERT INTO audit_log (
            audit_msg
        )
        VALUES (
            $1
        )
    `

	msg := string(p)

	_, err = db.Exec(query, msg)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}
