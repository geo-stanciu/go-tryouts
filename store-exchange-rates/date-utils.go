package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	ISODate           string = "2006-01-02"
	ISODateTime       string = "2006-01-02 15:04:05"
	ISODateTimestamp  string = "2006-01-02 15:04:05.000"
	ISODateTimeZ      string = "2006-01-02 15:04:05Z07:00"
	ISODateTimestampZ string = "2006-01-02 15:04:05.000Z07:00"
	DMY               string = "02/01/2006"
	DMYTime           string = "02/01/2006 15:04:05"
	UTCDateTime       string = "UTC"
	UTCDateTimestamp  string = "UTCTimestamp"
	DateOffset        string = "Z07:00"
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
	case ISODate, ISODateTime, ISODateTimestamp, ISODateTimeZ, ISODateTimestampZ, DMY, DMYTime:
		return val.Format(format)
	case UTCDateTime:
		return val.UTC().Format(ISODateTimeZ)
	case UTCDateTimestamp:
		return val.UTC().Format(ISODateTimestampZ)
	default:
		return ""
	}
}

func string2date(sval string, format string) (time.Time, error) {
	switch format {
	case ISODate, ISODateTime, ISODateTimestamp, ISODateTimeZ, ISODateTimestampZ, DMY, DMYTime, DateOffset:
		t, err := time.Parse(format, sval)
		if err != nil {
			return time.Now(), err
		}
		return t, nil
	default:
		return time.Now(), fmt.Errorf("Unknown datetime format \"%s\"", format)
	}
}

func server2ClientDmy(r *http.Request, serverTime time.Time) string {
	t := server2ClientLocal(r, serverTime)
	return date2string(t, DMY)
}

func server2ClientDmyTime(r *http.Request, serverTime time.Time) string {
	t := server2ClientLocal(r, serverTime)
	return date2string(t, DMYTime)
}

func server2ClientLocal(r *http.Request, serverTime time.Time) time.Time {
	timeOffset := 0

	cookie, err := r.Cookie("time_zone_offset")
	if err != nil && err != http.ErrNoCookie {
		return serverTime.UTC()
	} else if err == http.ErrNoCookie {
		timeOffset = 0
	} else {
		timeOffset = string2int(cookie.Value)
	}

	return serverTime.UTC().Add(time.Duration(-1*timeOffset) * time.Minute)
}
