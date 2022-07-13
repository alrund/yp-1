package migrations

import "database/sql"

func UpUniqueURLIndex(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS urls_url_uindex ON urls (url);`)
	if err != nil {
		return err
	}

	return nil
}

func DownUniqueURLIndex(tx *sql.Tx) error {
	_, err := tx.Exec("DROP INDEX IF EXISTS urls_url_uindex;")
	if err != nil {
		return err
	}

	return nil
}
