package storage

import (
	"log"
	"os"
	"testing"
	"time"

	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestStorageFileName = "test_storage"

func TestFileGetToken(t *testing.T) {
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
			`{"url":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				tokenValue: "yyy",
			},
			"yyy",
			false,
		},
		{
			"fail",
			`{"url":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				tokenValue: "zzz",
			},
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage := &File{
				FileName: TestStorageFileName,
			}
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
			`{"url":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				url: "url",
			},
			"yyy",
			false,
		},
		{
			"fail",
			`{"url":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
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
			storage := &File{
				FileName: TestStorageFileName,
			}
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

func TestFileGetURL(t *testing.T) {
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
			`{"url":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				tokenValue: "xxx",
			},
			"url",
			false,
		},
		{
			"fail",
			`{"url":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				tokenValue: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTestData(tt.storageState)
			defer clearTestData()
			storage := &File{
				FileName: TestStorageFileName,
			}
			got, err := storage.GetURL(tt.args.tokenValue)
			if tt.want != "" {
				require.NotNil(t, got)
				assert.Equal(t, tt.want, got)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
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
			`{"url":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				tokenValue: "xxx",
			},
			true,
			false,
		},
		{
			"success not found",
			`{"url":{"Value":"yyy","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
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
			storage := &File{
				FileName: TestStorageFileName,
			}
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
			`{"url":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
			args{
				url: "url",
			},
			true,
			false,
		},
		{
			"success not found",
			`{"url":{"Value":"xxx","Expire":"2022-06-13T20:45:35.857891406+03:00"}}`,
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
			storage := &File{
				FileName: TestStorageFileName,
			}
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
		url   string
		token *tkn.Token
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
				url: "url",
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
			defer clearTestData()
			storage := &File{
				FileName: TestStorageFileName,
			}
			err := storage.Set(tt.args.url, tt.args.token)
			state, storageErr := storage.restoreState()
			if storageErr != nil {
				log.Fatal(storageErr)
			}
			assert.Equal(t, tt.want, state[tt.args.url].Value)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewFile(TestStorageFileName)
			if err != nil {
				log.Fatal(err)
			}
			assert.Equal(t, tt.want, storage)
		})
	}
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

func clearTestData() {
	err := os.Remove(TestStorageFileName)
	if err != nil {
		log.Fatal(err)
	}
}
