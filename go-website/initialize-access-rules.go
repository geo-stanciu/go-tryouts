package main

import (
	"database/sql"

	"github.com/geo-stanciu/go-utils/utils"
)

const (
	allOtherRequests string = "--all-rest--"
)

type menuName struct {
	language string
	name     string
}

type menu struct {
	requestURL string
	name       []menuName
	roles      []userRole
}

func addGetRequestsAccessRules(tx *sql.Tx) error {
	menus := []*menu{
		{"index",
			[]menuName{{"EN", "Index"}},
			[]userRole{{"Member"}},
		},
		{"users",
			[]menuName{{"EN", "Users"}},
			[]userRole{{"Administrator"}},
		},
		{"about",
			[]menuName{{"EN", "About"}},
			[]userRole{{"Member"}},
		},
		{"login",
			[]menuName{{"EN", "Login"}},
			[]userRole{{"All"}},
		},
		{"register",
			[]menuName{{"EN", "Register"}},
			[]userRole{{"All"}},
		},
		{"change-password",
			[]menuName{{"EN", "Change Password"}},
			[]userRole{{"Member"}},
		},
	}

	err := setAccessRules(tx, "GET", menus)
	return err
}

func addPostRequestsAccessRules(tx *sql.Tx) error {
	menus := []*menu{
		{"login",
			[]menuName{{"EN", "Login"}},
			[]userRole{{"All"}},
		},
		{"logout",
			[]menuName{{"EN", "Logout"}},
			[]userRole{{"All"}},
		},
		{"register",
			[]menuName{{"EN", "Register"}},
			[]userRole{{"All"}},
		},
	}

	err := setAccessRules(tx, "POST", menus)
	return err
}

func addGeneralMemberRequestsAccessRules(tx *sql.Tx) error {
	menus := []*menu{
		{allOtherRequests,
			[]menuName{},
			[]userRole{{"Member"}},
		},
	}

	err := setAccessRules(tx, "All", menus)
	return err
}

func setAccessRules(tx *sql.Tx, reqType string, menus []*menu) error {
	var found bool
	var requestID int32
	var err error

	var pq *utils.PreparedQuery

	for _, m := range menus {
		if m.requestURL == allOtherRequests {
			pq = dbUtils.PQuery(`
				SELECT request_id
				  FROM request
				EXCEPT
				SELECT request_id
				  FROM request_role
			`)

			// kinda ridiculos
			// apparently on Postgres you need to close
			// your previous fetch on a transaction
			// before you can work again on it
			// for now: read id's in memory and run the next query on array
			// references: https://github.com/lib/pq/issues/81
			//             https://github.com/lib/pq/issues/635
			var reqId []int32
			err = dbUtils.ForEachRowTx(tx, pq, func(row *sql.Rows, sc *utils.SQLScan) error {
				err = row.Scan(&requestID)
				if err != nil {
					return err
				}

				reqId = append(reqId, requestID)
				return nil
			})
			if err != nil {
				return err
			}

			for _, req := range reqId {
				for _, r := range m.roles {
					err = addRequest2Role(tx, req, r.role)
					if err != nil {
						return err
					}
				}
			}

		} else {
			pq = dbUtils.PQuery(`
				SELECT request_id
				FROM request
				WHERE request_url = ?
				AND request_type = ?
			`, m.requestURL,
				reqType)

			err = tx.QueryRow(pq.Query, pq.Args...).Scan(&requestID)
			if err != nil {
				return err
			}

			for _, n := range m.name {
				pq := dbUtils.PQuery(`
					SELECT CASE WHEN EXISTS (
						SELECT 1
						FROM request_name
						WHERE request_id = ?
						AND language = ?
					) THEN 1 ELSE 0 END FROM dual
				`, requestID,
					n.language)

				err := tx.QueryRow(pq.Query, pq.Args...).Scan(&found)
				if err != nil {
					return err
				}

				if found {
					continue
				}

				pq = dbUtils.PQuery(`
					INSERT INTO request_name (
						request_id,
						language,
						name
					)
					VALUES (?, ?, ?)
				`, requestID,
					n.language,
					n.name)

				_, err = dbUtils.ExecTx(tx, pq)
				if err != nil {
					return err
				}
			}

			for _, r := range m.roles {
				err = addRequest2Role(tx, requestID, r.role)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func addRequest2Role(tx *sql.Tx, requestID int32, role string) error {
	var roleID int32
	var found bool

	pq := dbUtils.PQuery(`
		SELECT role_id
		FROM role
		WHERE loweredrole = lower(?)
	`, role)

	err := tx.QueryRow(pq.Query, pq.Args...).Scan(&roleID)
	if err != nil {
		return err
	}

	pq = dbUtils.PQuery(`
		SELECT CASE WHEN EXISTS (
			SELECT 1
			FROM request_role
			WHERE role_id = ?
			AND request_id = ?
		) THEN 1 ELSE 0 END FROM dual
	`, roleID,
		requestID)

	err = tx.QueryRow(pq.Query, pq.Args...).Scan(&found)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	pq = dbUtils.PQuery(`
		INSERT INTO request_role (
			role_id,
			request_id
		)
		VALUES (?, ?)
	`, roleID,
		requestID)

	_, err = dbUtils.ExecTx(tx, pq)
	if err != nil {
		return err
	}

	return nil
}
