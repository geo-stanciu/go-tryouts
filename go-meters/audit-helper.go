package main

type AuditLog struct {
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

	_, err = db.Exec(query, msg)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}
