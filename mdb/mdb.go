package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id int64 `json:"id"`
	Email string `json:"email"`
	ConfirmedAt *time.Time `json:"confirmed_at"`
	OptOut bool `json:"opt_out"`
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, email TEXT NOT NULL UNIQUE, confirmed_at INTEGER, opt_out BOOLEAN);")
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			// Code 1 if table already exists
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmedAt int64
	var optOut bool

	err := row.Scan(&id, &email, &confirmedAt, &optOut)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	t := time.Unix(confirmedAt, 0)													
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: optOut}, nil
}	

func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`INSERT INTO
		emails(email, confirmed_at, opt_out)
		VALUES(?, 0, false);
	`, email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query("SELECT * FROM emails WHERE email = ?;", email)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	// db.Query keeps the connection open, so we need to close it
	defer rows.Close()

	for rows.Next() {
		return emailEntryFromRow(rows)
	}

	return nil, nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()
	_, err := db.Exec(`
		INSERT INTO
		emails(email, confirmed_at, opt_out)
		VALUES(?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET confirmed_at = ?, opt_out = ?;
	`, entry.Email, t, entry.OptOut, t, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil;
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`UPDATE emails SET opt_out=true WHERE email=?`, email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type GetEmailBatchQueryParams struct {
	Page int
	Count int
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	rows, err := db.Query(`
		SELECT * FROM emails
		WHERE opt_out = false
		ORDER BY id ASC
		LIMIT ? OFFSET ?;
	`, params.Count, params.Count * (params.Page - 1))

	if err != nil {
		log.Println(err)
		return nil, err
	}
	// db.Query keeps the connection open, so we need to close it
	defer rows.Close()

	entries := make([]EmailEntry, 0, params.Count)
	for rows.Next() {
		entry, err := emailEntryFromRow(rows)
		if err != nil {
			return nil, err
		}

		entries = append(entries, *entry)
	}

	return entries, nil
}