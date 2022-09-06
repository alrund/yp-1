package storage

import (
	"testing"
	"time"

	tkn "github.com/alrund/yp-1/internal/app/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    string
		wantErr bool
	}{
		{
			"success",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				tokenValue: "xxx",
			},
			"yyy",
			false,
		},
		{
			"fail",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				tokenValue: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.GetToken(tt.args.tokenValue)
			if tt.want != "" {
				assert.Equal(t, tt.want, got.Value)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestGetTokenByURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    string
		wantErr bool
	}{
		{
			"success",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				url: "url",
			},
			"yyy",
			false,
		},
		{
			"fail",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				url: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.GetTokenByURL(tt.args.url)
			if tt.want != "" {
				assert.Equal(t, tt.want, got.Value)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestGetTokensByUserID(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    string
		wantErr bool
	}{
		{
			"success",
			&Map{
				userID2tokenValue: map[string][]string{"XXX-YYY-ZZZ": {"xxx"}},
				url2tokenValue:    map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL:    "url",
						UserID: "XXX-YYY-ZZZ",
					},
				},
			},
			args{
				userID: "XXX-YYY-ZZZ",
			},
			"yyy",
			false,
		},
		{
			"fail",
			&Map{
				userID2tokenValue: map[string][]string{"XXX-YYY-ZZZ": {"xxx"}},
				url2tokenValue:    map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL:    "url",
						UserID: "XXX-YYY-ZZZ",
					},
				},
			},
			args{
				userID: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.GetTokensByUserID(tt.args.userID)
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

func TestGetURL(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    string
		wantErr bool
	}{
		{
			"success",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				tokenValue: "xxx",
			},
			"url",
			false,
		},
		{
			"fail",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				tokenValue: "zzz",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.GetURL(tt.args.tokenValue)
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestHasToken(t *testing.T) {
	type args struct {
		tokenValue string
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				tokenValue: "xxx",
			},
			true,
			false,
		},
		{
			"success not found",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				tokenValue: "zzz",
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.HasToken(tt.args.tokenValue)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestHasURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    bool
		wantErr bool
	}{
		{
			"success",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				url: "url",
			},
			true,
			false,
		},
		{
			"success not found",
			&Map{
				url2tokenValue: map[string]string{"url": "xxx"},
				tokenValue2composite: map[string]*composite{
					"xxx": {
						Token: &tkn.Token{
							Value:  "yyy",
							Expire: time.Now().Add(tkn.LifeTime),
						},
						URL: "url",
					},
				},
			},
			args{
				url: "zzz",
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.storage.HasURL(tt.args.url)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestSet(t *testing.T) {
	type args struct {
		userID string
		url    string
		token  *tkn.Token
	}
	tests := []struct {
		name    string
		storage *Map
		args    args
		want    string
		wantErr bool
	}{
		{
			"success",
			&Map{
				userID2tokenValue:    make(map[string][]string),
				url2tokenValue:       map[string]string{},
				tokenValue2composite: map[string]*composite{},
			},
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
			err := tt.storage.Set(tt.args.userID, tt.args.url, tt.args.token)
			assert.Equal(t, tt.want, tt.storage.url2tokenValue[tt.args.url])
			assert.Equal(t, tt.want, tt.storage.userID2tokenValue[tt.args.userID][0])
			if tt.wantErr {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestNewMapStorage(t *testing.T) {
	tests := []struct {
		name string
		want *Map
	}{
		{
			name: "success",
			want: &Map{
				userID2tokenValue:    make(map[string][]string),
				url2tokenValue:       make(map[string]string),
				tokenValue2composite: make(map[string]*composite),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewMap())
		})
	}
}

func BenchmarkMapGet(b *testing.B) {
	storage := &Map{
		userID2tokenValue: map[string][]string{"XXX-YYY-ZZZ": {"xxx"}},
		url2tokenValue:    map[string]string{"url": "xxx"},
		tokenValue2composite: map[string]*composite{
			"xxx": {
				Token: &tkn.Token{
					Value:  "yyy",
					Expire: time.Now().Add(tkn.LifeTime),
				},
				URL:    "url",
				UserID: "XXX-YYY-ZZZ",
			},
		},
	}
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
