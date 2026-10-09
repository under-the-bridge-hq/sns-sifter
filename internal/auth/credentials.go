package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ErrNotFound は認証情報ファイルが存在しないことを示す。
var ErrNotFound = errors.New("認証情報ファイルが見つかりません")

type Credentials struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
	UserID            string `json:"user_id"`
	ScreenName        string `json:"screen_name"`
}

// Path は認証情報の保存先を返す。SIFTER_CREDENTIALS、無ければ ~/.sifter/credentials.json。
func Path() (string, error) {
	if p := os.Getenv("SIFTER_CREDENTIALS"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".sifter", "credentials.json"), nil
}

// Load は認証情報を読む。ファイルが無ければ ErrNotFound を返す。
func Load() (*Credentials, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s の解析に失敗: %w", p, err)
	}
	return &c, nil
}

// Save は認証情報を保存する (ディレクトリ 0700 / ファイル 0600)。保存先パスを返す。
func Save(c *Credentials) (string, error) {
	p, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return "", err
	}
	if err := os.Chmod(p, 0o600); err != nil {
		return "", err
	}
	return p, nil
}
