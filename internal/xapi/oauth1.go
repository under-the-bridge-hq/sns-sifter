package xapi

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// OAuth1Signer は OAuth 1.0a (HMAC-SHA1, RFC 5849) の署名とヘッダ生成を行う。
type OAuth1Signer struct {
	ConsumerKey    string
	ConsumerSecret string
	Token          string
	TokenSecret    string

	// テスト用に注入可能。nil なら既定実装を使う。
	Nonce func() string
	Now   func() time.Time
}

// PercentEncode は RFC 3986 の unreserved 以外を %XX (大文字) にエンコードする。
func PercentEncode(s string) string {
	const hexUpper = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~' {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hexUpper[c>>4])
		b.WriteByte(hexUpper[c&0x0f])
	}
	return b.String()
}

// EncodeQuery は url.Values を PercentEncode で key=value&... に直列化する (キー順)。
func EncodeQuery(v url.Values) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		for _, val := range v[k] {
			parts = append(parts, PercentEncode(k)+"="+PercentEncode(val))
		}
	}
	return strings.Join(parts, "&")
}

type kv struct{ k, v string }

func sortedEncoded(params []kv) []kv {
	enc := make([]kv, len(params))
	for i, p := range params {
		enc[i] = kv{PercentEncode(p.k), PercentEncode(p.v)}
	}
	sort.Slice(enc, func(i, j int) bool {
		if enc[i].k != enc[j].k {
			return enc[i].k < enc[j].k
		}
		return enc[i].v < enc[j].v
	})
	return enc
}

// SignatureBaseString は署名ベース文字列を返す。
// params は query / form body / oauth_* を含む全パラメータ (未エンコード)。
func SignatureBaseString(method, baseURL string, params url.Values) string {
	var list []kv
	for k, vs := range params {
		for _, v := range vs {
			list = append(list, kv{k, v})
		}
	}
	enc := sortedEncoded(list)
	parts := make([]string, len(enc))
	for i, p := range enc {
		parts[i] = p.k + "=" + p.v
	}
	return strings.ToUpper(method) + "&" + PercentEncode(baseURL) + "&" + PercentEncode(strings.Join(parts, "&"))
}

// Sign は署名ベース文字列から oauth_signature (base64) を計算する。
func (s *OAuth1Signer) Sign(baseString string) string {
	key := PercentEncode(s.ConsumerSecret) + "&" + PercentEncode(s.TokenSecret)
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(baseString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (s *OAuth1Signer) nonce() string {
	if s.Nonce != nil {
		return s.Nonce()
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func (s *OAuth1Signer) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// oauthParams は署名対象に含める oauth_* パラメータを返す (signature を除く)。
func (s *OAuth1Signer) oauthParams(extra map[string]string) url.Values {
	p := url.Values{}
	p.Set("oauth_consumer_key", s.ConsumerKey)
	p.Set("oauth_nonce", s.nonce())
	p.Set("oauth_signature_method", "HMAC-SHA1")
	p.Set("oauth_timestamp", strconv.FormatInt(s.now().Unix(), 10))
	if s.Token != "" {
		p.Set("oauth_token", s.Token)
	}
	p.Set("oauth_version", "1.0")
	for k, v := range extra {
		p.Set(k, v)
	}
	return p
}

// AuthorizationHeader は Authorization ヘッダ値を返す。
// baseURL は query を除いた URL、query / form は実際に送るパラメータ、
// extraOAuth は oauth_callback / oauth_verifier など追加の oauth_* パラメータ。
func (s *OAuth1Signer) AuthorizationHeader(method, baseURL string, query, form url.Values, extraOAuth map[string]string) string {
	oauth := s.oauthParams(extraOAuth)
	all := url.Values{}
	for _, src := range []url.Values{query, form, oauth} {
		for k, vs := range src {
			for _, v := range vs {
				all.Add(k, v)
			}
		}
	}
	oauth.Set("oauth_signature", s.Sign(SignatureBaseString(method, baseURL, all)))

	keys := make([]string, 0, len(oauth))
	for k := range oauth {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = PercentEncode(k) + `="` + PercentEncode(oauth.Get(k)) + `"`
	}
	return "OAuth " + strings.Join(parts, ", ")
}
