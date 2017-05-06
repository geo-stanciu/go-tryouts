package main

import (
	"sync"
	"time"
)

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

func Log(isErr bool, err error, msgType string, msg string, details ...interface{}) {
	fields := make(map[string]interface{})

	if len(msgType) > 0 {
		fields["msg_type"] = msgType
	}

	if isErr {
		fields["status"] = "failed"
	} else {
		fields["status"] = "successful"
	}

	if details != nil {
		var key string

		for i, detail := range details {
			if i%2 == 0 {
				key = detail.(string)
			} else {
				fields[key] = detail
			}
		}
	}

	hasKeys := false

	if len(fields) > 0 {
		hasKeys = true
	}

	if isErr {
		if hasKeys {
			if err != nil {
				log.WithError(err).WithFields(fields).Error(msg)
			} else {
				log.WithFields(fields).Error(msg)
			}
		} else {
			if err != nil {
				log.WithError(err).Error(msg)
			} else {
				log.Error(msg)
			}
		}
	} else {
		if hasKeys {
			if err != nil {
				log.WithError(err).WithFields(fields).Info(msg)
			} else {
				log.WithFields(fields).Info(msg)
			}
		} else {
			if err != nil {
				log.WithError(err).Info(msg)
			} else {
				log.Info(msg)
			}
		}
	}
}
