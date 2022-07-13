package migrations

import "database/sql"

func UpTokens(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS tokens
		(
			token VARCHAR(6)  NOT NULL PRIMARY KEY,
			expire INTEGER  NOT NULL
		);`,
	)
	if err != nil {
		return err
	}

	return nil
}

func DownTokens(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS tokens;")
	if err != nil {
		return err
	}

	return nil
}
