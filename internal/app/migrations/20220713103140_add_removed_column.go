package migrations

import "database/sql"

func UpAddRemovedColumn(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE tokens ADD IF NOT EXISTS removed BOOL DEFAULT FALSE NOT NULL;`)
	if err != nil {
		return err
	}

	return nil
}

func DownAddRemovedColumn(tx *sql.Tx) error {
	_, err := tx.Exec("ALTER TABLE TOKENS DROP COLUMN removed;")
	if err != nil {
		return err
	}

	return nil
}
