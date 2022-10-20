package storage

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestStorageFileName = "test_storage"

func TestRaceFileGetToken(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         string
		wantErr      bool
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "yyy",
			},
			"yyy",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			go func() {
				_ = storage.Set("UUU", "URL", &tkn.Token{
					Value:  "yyy",
					Expire: time.Now().Add(tkn.LifeTime),
				})
			}()
			got, err := storage.GetToken(tt.args.tokenValue)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got.Value)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFileGetToken(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         string
		wantErr      *error
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "yyy",
			},
			"yyy",
			nil,
		},
		{
			"fail",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "zzz",
			},
			"",
			&ErrTokenNotFound,
		},
		{
			"fail - no token",
			`{
							"url":{
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "zzz",
			},
			"",
			&ErrTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			got, err := storage.GetToken(tt.args.tokenValue)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got.Value)
			}
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, *tt.wantErr)
			}
		})
	}
}

func TestFileGetTokenByURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         string
		wantErr      bool
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				url: "url",
			},
			"yyy",
			false,
		},
		{
			"fail",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				url: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			got, err := storage.GetTokenByURL(tt.args.url)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got.Value)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFileGetTokensByUserID(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         string
		wantErr      bool
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				userID: "XXX-YYY-ZZZ",
			},
			"yyy",
			false,
		},
		{
			"fail",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				userID: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			got, err := storage.GetTokensByUserID(tt.args.userID)
			if tt.want != "" {
				require.NotNil(t, got)
				require.Greater(t, len(got), 0)
				assert.Equal(t, tt.want, got[0].Value)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFileGetURLsByUserID(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         URLpairs
		wantErr      *error
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				userID: "XXX-YYY-ZZZ",
			},
			URLpairs{
				OriginalURL: "url",
			},
			nil,
		},
		{
			"fail - incorrect userID",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				userID: "incorrect",
			},
			URLpairs{},
			&ErrTokenNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			urls, err := storage.GetURLsByUserID(tt.args.userID)
			if tt.want.OriginalURL != "" {
				require.NotNil(t, urls)
				assert.Equal(t, tt.want.OriginalURL, urls[0].OriginalURL)
			}
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, *tt.wantErr)
			}
		})
	}
}

func TestFileGetURL(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         string
		wantErr      *error
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "xxx",
			},
			"url",
			nil,
		},
		{
			"fail",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "zzz",
			},
			"",
			&ErrURLNotFound,
		},
		{
			"fail - no token",
			`{
							"url":{
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "zzz",
			},
			"",
			&ErrTokenNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			got, err := storage.GetURL(tt.args.tokenValue)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, *tt.wantErr)
			}
		})
	}
}

func TestFileHasToken(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         bool
		wantErr      bool
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "xxx",
			},
			true,
			false,
		},
		{
			"success not found",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				tokenValue: "zzz",
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			got, err := storage.HasToken(tt.args.tokenValue)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFileHasURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         bool
		wantErr      bool
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				url: "url",
			},
			true,
			false,
		},
		{
			"success not found",
			`{
							"url":{"Token":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				url: "zzz",
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			got, err := storage.HasURL(tt.args.url)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFileSet(t *testing.T) {
	type args struct {
		userID string
		url    string
		token  *tkn.Token
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
				userID: "XXX-YYY-ZZZ",
				url:    "url",
				token: &tkn.Token{
					Value:  "yyy",
					Expire: time.Now().Add(tkn.LifeTime),
				},
			},
			"yyy",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData("")
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			err = storage.Set(tt.args.userID, tt.args.url, tt.args.token)
			assert.NotNil(t, storage.state[tt.args.url].Token)
			assert.Equal(t, tt.want, storage.state[tt.args.url].Token.Value)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFileSetBatch(t *testing.T) {
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
				userID: "XXX-YYY-ZZZ",
				url2token: map[string]*tkn.Token{
					"url": {
						Value:  "yyy",
						Expire: time.Now().Add(tkn.LifeTime),
					},
					"url2": {
						Value:  "yyy2",
						Expire: time.Now().Add(tkn.LifeTime),
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData("")
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			err = storage.SetBatch(tt.args.userID, tt.args.url2token)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			for url, token := range tt.args.url2token {
				assert.NotNil(t, storage.state[url].Token)
				assert.Equal(t, token.Value, storage.state[url].Token.Value)
			}
		})
	}
}

func TestFileRemoveTokens(t *testing.T) {
	type args struct {
		userID      string
		tokenValues []string
	}
	tests := []struct {
		name         string
		storageState string
		args         args
		want         bool
		wantErr      bool
	}{
		{
			"success",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				userID:      "XXX-YYY-ZZZ",
				tokenValues: []string{"yyy"},
			},
			true,
			false,
		},
		{
			"fail - incorrect userID",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			args{
				userID:      "incorrect",
				tokenValues: []string{"yyy"},
			},
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			err = storage.RemoveTokens(tt.args.tokenValues, tt.args.userID)
			if tt.wantErr {
				require.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			if tt.want {
				assert.True(t, isTestDataContainsString("\"Removed\":true"))
			} else {
				assert.True(t, isTestDataContainsString("\"Removed\":false"))
			}
		})
	}
}

func TestFilePing(t *testing.T) {
	createTestData("")
	defer clearTestData()
	storage, err := NewFile(TestStorageFileName)
	require.NoError(t, err)
	assert.Nil(t, storage.Ping(context.Background()))
}

func TestFileGetURLCount(t *testing.T) {
	tests := []struct {
		name         string
		storageState string
		want         int
	}{
		{
			"success - empty",
			`{}`,
			0,
		},
		{
			"success - two",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			num, err := storage.GetURLCount()
			require.Nil(t, err)
			assert.Equal(t, tt.want, num)
		})
	}
}

func TestFileGetUserIDCount(t *testing.T) {
	tests := []struct {
		name         string
		storageState string
		want         int
	}{
		{
			"success - empty",
			`{}`,
			0,
		},
		{
			"success - two",
			`{
							"url":{"Token":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			require.NoError(t, err)
			num, err := storage.GetUserIDCount()
			require.Nil(t, err)
			assert.Equal(t, tt.want, num)
		})
	}
}

func TestNewFileStorage(t *testing.T) {
	tests := []struct {
		name string
		want *File
	}{
		{
			name: "success",
			want: &File{
				FileName: TestStorageFileName,
				state:    make(map[string]composite),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData("")
			defer clearTestData()
			storage, err := NewFile(TestStorageFileName)
			if err != nil {
				log.Fatal(err)
			}
			assert.Equal(t, tt.want, storage)
		})
	}
}

func BenchmarkFileGet(b *testing.B) {
	createTestData(`{
							"url":{"Token":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"},
							"URL":"url",
							"UserID":"XXX-YYY-ZZZ"}
						}`)
	defer clearTestData()
	storage, _ := NewFile(TestStorageFileName)
	b.Run("GetToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.GetToken("xxx")
		}
	})
	b.Run("GetTokenByURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.GetTokenByURL("url")
		}
	})
	b.Run("GetTokensByUserID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.GetTokensByUserID("XXX-YYY-ZZZ")
		}
	})
	b.Run("GetURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.GetURL("xxx")
		}
	})
	b.Run("HasToken", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.HasToken("xxx")
		}
	})
	b.Run("HasURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = storage.HasURL("url")
		}
	})
	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = storage.Set("XXX-YYY-ZZZ", "url", &tkn.Token{
				Value:  "yyy",
				Expire: time.Now().Add(tkn.LifeTime),
			})
		}
	})
}

func createTestData(testJSON string) {
	file, err := os.Create(TestStorageFileName)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write([]byte(testJSON))
	if err != nil {
		log.Fatal(err)
	}
}

func isTestDataContainsString(str string) bool {
	read, err := os.ReadFile(TestStorageFileName)
	if err != nil {
		log.Fatal(err)
	}

	return strings.Contains(string(read), str)
}

func clearTestData() {
	_ = os.Remove(TestStorageFileName)
}
