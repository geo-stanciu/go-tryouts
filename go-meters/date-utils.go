package main

import (
	"fmt"
	"time"
)

const (
	ISODate          string = "2006-01-02"
	ISODateTime      string = "2006-01-02T15:04:05Z07:00"
	ISODateTimestamp string = "2006-01-02T15:04:05.000Z07:00"
	UTCDateTime      string = "UTC"
	UTCDateTimestamp string = "UTCTimestamp"
)

func isISODate(sval string) bool {
	_, err := string2date(sval, ISODate)

	if err != nil {
		return false
	}

	return true
}

func isISODateTime(sval string) bool {
	_, err := string2date(sval, ISODateTime)

	if err != nil {
		return false
	}

	return true
}

func dateFromISODateTime(sval string) (time.Time, error) {
	return string2date(sval, ISODateTime)
}

func date2string(val time.Time, format string) string {
	switch format {
	case ISODate, ISODateTime, ISODateTimestamp:
		return val.Format(format)
	case UTCDateTime:
		return val.UTC().Format(ISODateTime)
	case UTCDateTimestamp:
		return val.UTC().Format(ISODateTimestamp)
	default:
		return ""
	}
}

func string2date(sval string, format string) (time.Time, error) {
	switch format {
	case ISODate, ISODateTime, ISODateTimestamp, UTCDateTime:
		t, err := time.Parse(format, sval)
		if err != nil {
			return time.Now(), err
		}
		return t, nil
	default:
		return time.Now(), fmt.Errorf("Unknown datetime format \"%s\"", format)
	}
}
