package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alrund/yp-1/internal/app/migrations"
	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v4/stdlib" // pgx
)

type DB struct {
	db *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	newDB := &DB{
		db: db,
	}

	err = newDB.migrations()
	if err != nil {
		return nil, err
	}

	return newDB, nil
}

func (d *DB) migrations() error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = migrations.UpTokens(tx)
	if err != nil {
		return err
	}

	err = migrations.UpUrls(tx)
	if err != nil {
		return err
	}

	err = migrations.UpUniqueURLIndex(tx)
	if err != nil {
		return err
	}

	err = migrations.UpAddRemovedColumn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) Set(userID, url string, token *tkn.Token) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = d.db.Exec("INSERT INTO tokens(token, expire) VALUES($1, $2)", token.Value, token.Expire.Unix())
	if err != nil {
		return err
	}

	_, err = d.db.Exec(
		"INSERT INTO urls(id, url, token, user_id) VALUES($1, $2, $3, $4)",
		uuid.NewString(),
		url,
		token.Value,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) SetBatch(userID string, url2token map[string]*tkn.Token) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	tokensStmt, err := d.db.Prepare("INSERT INTO tokens(token, expire) VALUES($1, $2)")
	if err != nil {
		return err
	}
	defer tokensStmt.Close()

	urlsStmt, err := d.db.Prepare("INSERT INTO urls(id, url, token, user_id) VALUES($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer urlsStmt.Close()

	for url, token := range url2token {
		_, err = tokensStmt.Exec(token.Value, token.Expire.Unix())
		if err != nil {
			return err
		}

		_, err = urlsStmt.Exec(uuid.NewString(), url, token.Value, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *DB) RemoveTokens(tokenValues []string, userID string) error {
	num := len(tokenValues)

	valPhs := ""
	vals := make([]interface{}, 0, num)
	for i, tokenValue := range tokenValues {
		valPhs += "($" + strconv.Itoa(i+1) + "),"
		vals = append(vals, tokenValue)
	}
	vals = append(vals, userID)

	stmt, err := d.db.Prepare(fmt.Sprintf(
		"UPDATE tokens SET removed=true FROM (VALUES %s) AS tmp (token) "+
			"WHERE tokens.token=tmp.token AND tokens.token IN (SELECT token FROM urls WHERE user_id=$%d)",
		strings.TrimRight(valPhs, ","), num+1),
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) GetToken(tokenValue string) (*tkn.Token, error) {
	var value string
	var timestamp int64
	var removed bool
	err := d.db.QueryRow(
		"SELECT token, expire, removed FROM tokens WHERE token = $1", tokenValue,
	).Scan(&value, &timestamp, &removed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	return &tkn.Token{
		Value:   value,
		Expire:  time.Unix(timestamp, 0),
		Removed: removed,
	}, nil
}

func (d *DB) GetTokenByURL(url string) (*tkn.Token, error) {
	var value string
	var timestamp int64
	var removed bool
	err := d.db.QueryRow(
		"SELECT t.token, t.expire, t.removed FROM tokens t, urls u WHERE u.token = t.token AND u.url = $1", url,
	).Scan(&value, &timestamp, &removed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	return &tkn.Token{
		Value:   value,
		Expire:  time.Unix(timestamp, 0),
		Removed: removed,
	}, nil
}

func (d *DB) GetTokensByUserID(userID string) ([]*tkn.Token, error) {
	rows, err := d.db.Query(
		"SELECT t.token, t.expire, t.removed FROM tokens t, urls u WHERE u.token = t.token AND u.user_id = $1", userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	defer rows.Close()

	tokens := make([]*tkn.Token, 0)
	for rows.Next() {
		var value string
		var timestamp int64
		var removed bool
		err = rows.Scan(&value, &timestamp, &removed)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, &tkn.Token{
			Value:   value,
			Expire:  time.Unix(timestamp, 0),
			Removed: removed,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (d *DB) GetURL(tokenValue string) (string, error) {
	var url string
	err := d.db.QueryRow(
		"SELECT url FROM urls WHERE token = $1", tokenValue,
	).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}
		return "", err
	}

	return url, nil
}

func (d *DB) GetURLsByUserID(userID string) ([]URLpairs, error) {
	rows, err := d.db.Query(
		"SELECT url, token FROM urls WHERE user_id = $1", userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrURLNotFound
		}
		return nil, err
	}

	defer rows.Close()

	pairs := make([]URLpairs, 0)
	for rows.Next() {
		var url string
		var token string
		err = rows.Scan(&url, &token)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, URLpairs{
			OriginalURL: url,
			ShortURL:    token,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return pairs, nil
}

func (d *DB) HasURL(url string) (bool, error) {
	var u string
	err := d.db.QueryRow(
		"SELECT url FROM urls WHERE url = $1", url,
	).Scan(&u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return u != "", nil
}

func (d *DB) HasToken(tokenValue string) (bool, error) {
	var t string
	err := d.db.QueryRow(
		"SELECT token FROM tokens WHERE token = $1", tokenValue,
	).Scan(&t)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return t != "", nil
}

func (d *DB) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	return d.db.PingContext(ctx)
}
