package migrations

import "database/sql"

func UpUrls(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS urls
		(
			id UUID NOT NULL PRIMARY KEY,
			url VARCHAR(255)  NOT NULL,
			token VARCHAR(6)  NOT NULL CONSTRAINT urls_tokens_fk REFERENCES tokens,
			user_id UUID
		);`,
	)
	if err != nil {
		return err
	}

	return nil
}

func DownUrls(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE IF EXISTS urls;")
	if err != nil {
		return err
	}

	return nil
}
