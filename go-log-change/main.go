package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/geo-stanciu/go-utils/utils"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

var (
	appName    = "GoLogChange"
	appVersion = "0.0.0.1"
	log        = logrus.New()
	audit      = utils.AuditLog{}
	db         *sql.DB
	dbutl      *utils.DbUtils
	config     = Configuration{}
	currentDir string
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.DebugLevel

	dbutl = new(utils.DbUtils)
	currentDir = filepath.Dir(os.Args[0])
}

func main() {
	var err error
	var wg sync.WaitGroup

	cfgFile := fmt.Sprintf("%s/conf.json", currentDir)
	err = config.ReadFromFile(cfgFile)
	if err != nil {
		log.Println(err)
		return
	}

	err = dbutl.Connect2Database(&db, config.DbType, config.DbURL)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	audit.SetLogger(appName, appVersion, log, dbutl)
	audit.SetWaitGroup(&wg)
	defer audit.Close()

	mw := io.MultiWriter(os.Stdout, audit)
	log.Out = mw

	err = changeKeyName("add exchange rate", "data")
	if err != nil {
		log.Println(err)
		return
	}

	wg.Wait()
}

func changeKeyName(msgType string, oldKeyName string) error {
	tx, err := dbutl.BeginTransaction()
	if err != nil {
		return err
	}
	defer dbutl.Rollback(tx)

	mtype := fmt.Sprintf("{\"msg_type\": \"%s\"}", msgType)

	pq := dbutl.PQuery(`
		select audit_log_id, log_msg
	      from audit_log
		 where log_msg @> ?
		   and log_msg ?? ?
		   limit ?
	`, mtype,
		oldKeyName,
		1000)

	var messages []auditMsg

	err = dbutl.ForEachRowTx(tx, pq, func(row *sql.Rows, sc *utils.SQLScan) error {
		var msgID int64
		var msg string
		err = row.Scan(&msgID, &msg)
		if err != nil {
			return err
		}

		var old oldAddExchangeRateMsg
		r := strings.NewReader(msg)
		decoder := json.NewDecoder(r)

		err = decoder.Decode(&old)
		if err != nil {
			return err
		}

		var new newAddExchangeRateMsg
		new.Msg = old.Msg
		new.Date = old.Data
		new.Rate = old.Rate
		new.Time = old.Time
		new.Level = old.Level
		new.Status = old.Status
		new.Currency = old.Currency
		new.MsgType = old.MsgType

		aMsg := auditMsg{
			msgID: msgID,
			old:   &old,
			new:   &new,
		}

		messages = append(messages, aMsg)

		return nil
	})

	if err != nil {
		return err
	}

	pq = dbutl.PQuery(`
	    UPDATE audit_log SET log_msg = ? WHERE audit_log_id = ? 
	`)

	for _, msg := range messages {
		bmsg, err := json.Marshal(msg.new)
		if err != nil {
			return err
		}
		content := string(bmsg)

		pq.SetArg(0, content)
		pq.SetArg(1, msg.msgID)

		dbutl.ExecTx(tx, pq)

		audit.Log(nil,
			"adjust log - add exchange rate",
			"change log field names",
			"old", msg.old,
			"new", msg.new)
	}

	dbutl.Commit(tx)

	return nil
}

type auditMsg struct {
	msgID int64
	old   interface{}
	new   interface{}
}

type oldAddExchangeRateMsg struct {
	Msg      string    `json:"msg"`
	Data     string    `json:"data"`
	Rate     float64   `json:"rate"`
	Time     time.Time `json:"time"`
	Level    string    `json:"level"`
	Status   string    `json:"status"`
	Currency string    `json:"currency"`
	MsgType  string    `json:"msg_type"`
}

type newAddExchangeRateMsg struct {
	Msg      string    `json:"msg"`
	Date     string    `json:"date"`
	Rate     float64   `json:"rate"`
	Time     time.Time `json:"time"`
	Level    string    `json:"level"`
	Status   string    `json:"status"`
	Currency string    `json:"currency"`
	MsgType  string    `json:"msg_type"`
}
