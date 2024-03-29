package storage

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDbGetToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"token", "expire", "removed"}).
		AddRow("qwerty", time.Now().Add(tkn.LifeTime).Unix(), false)
	query := mock.ExpectQuery("^SELECT token, expire, removed FROM tokens WHERE token = (.+)")
	query.WithArgs("qwerty").WillReturnRows(rows)

	type args struct {
		tokenValue string
	}
	tests := []struct {
		name string
		args args
		want *tkn.Token
	}{
		{
			"success",
			args{
				tokenValue: "qwerty",
			},
			&tkn.Token{
				Value:   "qwerty",
				Removed: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetToken(tt.args.tokenValue)
			require.Nil(t, err)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Value, got.Value)
				assert.Equal(t, tt.want.Removed, got.Removed)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetTokenFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	query := mock.ExpectQuery("^SELECT token, expire, removed FROM tokens WHERE token = (.+)")
	query.WithArgs("zzz").WillReturnError(sql.ErrNoRows)

	type args struct {
		tokenValue string
	}
	tests := []struct {
		name    string
		args    args
		want    *tkn.Token
		wantErr error
	}{
		{
			"fail",
			args{
				tokenValue: "zzz",
			},
			nil,
			ErrTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetToken(tt.args.tokenValue)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Value, got.Value)
				assert.Equal(t, tt.want.Removed, got.Removed)
			}
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetTokenByURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"token", "expire", "removed"}).
		AddRow("qwerty", time.Now().Add(tkn.LifeTime).Unix(), false)
	mock.ExpectQuery("^SELECT t.token, t.expire, t.removed FROM tokens t, urls u " +
		"WHERE u.token = t.token AND u.url = (.+)").
		WithArgs("http://ya.ru").
		WillReturnRows(rows)

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *tkn.Token
		wantErr bool
	}{
		{
			"success",
			args{
				url: "http://ya.ru",
			},
			&tkn.Token{
				Value:   "qwerty",
				Removed: false,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetTokenByURL(tt.args.url)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Value, got.Value)
				assert.Equal(t, tt.want.Removed, got.Removed)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetTokenByURLFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("^SELECT t.token, t.expire, t.removed FROM tokens t, urls u " +
		"WHERE u.token = t.token AND u.url = (.+)").
		WithArgs("http://ya.ru").
		WillReturnError(sql.ErrNoRows)

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    *tkn.Token
		wantErr error
	}{
		{
			"fail",
			args{
				url: "http://ya.ru",
			},
			nil,
			ErrTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetTokenByURL(tt.args.url)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Value, got.Value)
				assert.Equal(t, tt.want.Removed, got.Removed)
			}
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, ErrTokenNotFound)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetTokensByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"token", "expire", "removed"}).
		AddRow("qwerty", time.Now().Add(tkn.LifeTime).Unix(), false)
	mock.ExpectQuery("^SELECT t.token, t.expire, t.removed FROM tokens t, urls u " +
		"WHERE u.token = t.token AND u.user_id = (.+)").
		WithArgs("591c1645-e1bb-4f64-bf8e-7eef7e5bff94").
		WillReturnRows(rows)

	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    *tkn.Token
		wantErr bool
	}{
		{
			"success",
			args{
				userID: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
			},
			&tkn.Token{
				Value:   "qwerty",
				Removed: false,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetTokensByUserID(tt.args.userID)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Value, got[0].Value)
				assert.Equal(t, tt.want.Removed, got[0].Removed)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetTokensByUserIDFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("^SELECT t.token, t.expire, t.removed FROM tokens t, urls u " +
		"WHERE u.token = t.token AND u.user_id = (.+)").
		WithArgs("591c1645-e1bb-4f64-bf8e-7eef7e5bff94").
		WillReturnError(sql.ErrNoRows)

	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    *tkn.Token
		wantErr error
	}{
		{
			"fail",
			args{
				userID: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
			},
			nil,
			ErrTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetTokensByUserID(tt.args.userID)
			if tt.want != nil {
				assert.Equal(t, tt.want.Value, got[0].Value)
				assert.Equal(t, tt.want.Removed, got[0].Removed)
			}
			assert.ErrorIs(t, err, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"url", "token"}).
		AddRow("http://ya.ru", "qwerty")
	mock.ExpectQuery("^SELECT url FROM urls WHERE token = (.+)").
		WithArgs("qwerty").
		WillReturnRows(rows)

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"success",
			args{
				url: "qwerty",
			},
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetURL(tt.args.url)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetURLFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("^SELECT url FROM urls WHERE token = (.+)").
		WithArgs("qwerty").
		WillReturnError(sql.ErrNoRows)

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			"fail",
			args{
				url: "qwerty",
			},
			"",
			ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetURL(tt.args.url)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
			assert.ErrorIs(t, err, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetURLsByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"url", "token"}).
		AddRow("http://ya.ru", "qwerty")
	mock.ExpectQuery("^SELECT url, token FROM urls WHERE user_id = (.+)").
		WithArgs("591c1645-e1bb-4f64-bf8e-7eef7e5bff94").
		WillReturnRows(rows)

	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    *URLpairs
		wantErr bool
	}{
		{
			"success",
			args{
				userID: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
			},
			&URLpairs{
				ShortURL:    "qwerty",
				OriginalURL: "http://ya.ru",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetURLsByUserID(tt.args.userID)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.ShortURL, got[0].ShortURL)
				assert.Equal(t, tt.want.OriginalURL, got[0].OriginalURL)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetURLsByUserIDFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("^SELECT url, token FROM urls WHERE user_id = (.+)").
		WithArgs("591c1645-e1bb-4f64-bf8e-7eef7e5bff94").
		WillReturnError(sql.ErrNoRows)

	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		args    args
		want    *URLpairs
		wantErr error
	}{
		{
			"fail",
			args{
				userID: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
			},
			nil,
			ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetURLsByUserID(tt.args.userID)
			if tt.want != nil {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.ShortURL, got[0].ShortURL)
				assert.Equal(t, tt.want.OriginalURL, got[0].OriginalURL)
			}
			assert.ErrorIs(t, err, tt.wantErr)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbHasURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"url"}).
		AddRow("http://ya.ru")
	mock.ExpectQuery("^SELECT url FROM urls WHERE url = (.+)").
		WithArgs("http://ya.ru").
		WillReturnRows(rows)

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success - true",
			args{
				url: "http://ya.ru",
			},
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.HasURL(tt.args.url)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbHasURLFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("^SELECT url FROM urls WHERE url = (.+)").
		WithArgs("http://google.com").
		WillReturnError(sql.ErrNoRows)

	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success - false",
			args{
				url: "http://google.com",
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.HasURL(tt.args.url)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbHasToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"token"}).
		AddRow("qwerty")
	mock.ExpectQuery("^SELECT token FROM tokens WHERE token = (.+)").
		WithArgs("qwerty").
		WillReturnRows(rows)

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success - true",
			args{
				token: "qwerty",
			},
			true,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.HasToken(tt.args.token)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbHasTokenFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("^SELECT token FROM tokens WHERE token = (.+)").
		WithArgs("qwerty").
		WillReturnError(sql.ErrNoRows)

	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success - false",
			args{
				token: "qwerty",
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.HasToken(tt.args.token)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbRemoveTokens(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("UPDATE tokens SET removed=true").
		ExpectExec().
		WithArgs("qwerty", "591c1645-e1bb-4f64-bf8e-7eef7e5bff94").
		WillReturnResult(sqlmock.NewResult(0, 1))

	type args struct {
		tokenValues []string
		userID      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"success - true",
			args{
				tokenValues: []string{"qwerty"},
				userID:      "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
			},
			false,
		},
		{
			"success - false",
			args{
				tokenValues: []string{"xxx"},
				userID:      "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			err := storage.RemoveTokens(tt.args.tokenValues, tt.args.userID)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbPing(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			"success",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			err := storage.Ping(context.Background())
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestDbSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	expireTime := time.Now().Add(tkn.LifeTime)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO tokens(token, expire) VALUES($1, $2)")).
		WithArgs("qwerty", expireTime.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO urls(id, url, token, user_id) VALUES($1, $2, $3, $4)")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	type args struct {
		userID, url string
		token       *tkn.Token
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"success",
			args{
				userID: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
				url:    "http://ya.ru",
				token: &tkn.Token{
					Value:  "qwerty",
					Expire: expireTime,
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			err := storage.Set(tt.args.userID, tt.args.url, tt.args.token)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbSetBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	expireTime := time.Now().Add(tkn.LifeTime)

	mock.ExpectBegin()
	tokensStmt := mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO tokens(token, expire) VALUES($1, $2)"))
	urlsStmt := mock.ExpectPrepare(
		regexp.QuoteMeta("INSERT INTO urls(id, url, token, user_id) VALUES($1, $2, $3, $4)"),
	)
	tokensStmt.ExpectExec().
		WithArgs("qwerty", expireTime.Unix()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	urlsStmt.ExpectExec().
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	type args struct {
		userID    string
		url2token map[string]*tkn.Token
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"success",
			args{
				userID: "591c1645-e1bb-4f64-bf8e-7eef7e5bff94",
				url2token: map[string]*tkn.Token{
					"http://ya.ru": {
						Value:  "qwerty",
						Expire: expireTime,
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			err := storage.SetBatch(tt.args.userID, tt.args.url2token)
			if tt.wantErr {
				assert.NotNil(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetURLCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"num"}).
		AddRow("3")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(DISTINCT token) as num FROM url")).
		WillReturnRows(rows)

	tests := []struct {
		name string
		want int
	}{
		{
			"success",
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetURLCount()

			require.Nil(t, err)
			assert.Equal(t, tt.want, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetURLCountFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(DISTINCT token) as num FROM url")).
		WillReturnError(sql.ErrNoRows)

	tests := []struct {
		name string
		want int
	}{
		{
			"no rows",
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetURLCount()

			require.Nil(t, err)
			assert.Equal(t, tt.want, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetUserIDCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"num"}).
		AddRow("3")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(DISTINCT user_id) as num FROM url")).
		WillReturnRows(rows)

	tests := []struct {
		name string
		want int
	}{
		{
			"success",
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetUserIDCount()

			require.Nil(t, err)
			assert.Equal(t, tt.want, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDbGetUserIDCountFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(DISTINCT user_id) as num FROM url")).
		WillReturnError(sql.ErrNoRows)

	tests := []struct {
		name string
		want int
	}{
		{
			"no rows",
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &DB{db: db}
			got, err := storage.GetUserIDCount()

			require.Nil(t, err)
			assert.Equal(t, tt.want, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
