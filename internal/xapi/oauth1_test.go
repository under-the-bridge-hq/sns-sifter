package xapi

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func vectorParams() url.Values {
	p := url.Values{}
	p.Set("include_entities", "true")
	p.Set("status", "Hello Ladies + Gentlemen, a signed OAuth request!")
	p.Set("oauth_consumer_key", "xvz1evFS4wEEPTGEFPHBog")
	p.Set("oauth_nonce", "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg")
	p.Set("oauth_signature_method", "HMAC-SHA1")
	p.Set("oauth_timestamp", "1318622958")
	p.Set("oauth_token", "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb")
	p.Set("oauth_version", "1.0")
	return p
}

const vectorBase = "POST&https%3A%2F%2Fapi.x.com%2F1.1%2Fstatuses%2Fupdate.json&include_entities%3Dtrue%26oauth_consumer_key%3Dxvz1evFS4wEEPTGEFPHBog%26oauth_nonce%3DkYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1318622958%26oauth_token%3D370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb%26oauth_version%3D1.0%26status%3DHello%2520Ladies%2520%252B%2520Gentlemen%252C%2520a%2520signed%2520OAuth%2520request%2521"

func vectorSigner() *OAuth1Signer {
	return &OAuth1Signer{
		ConsumerKey:    "xvz1evFS4wEEPTGEFPHBog",
		ConsumerSecret: "kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
		Token:          "370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
		TokenSecret:    "LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE",
		Nonce:          func() string { return "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg" },
		Now:            func() time.Time { return time.Unix(1318622958, 0) },
	}
}

func TestSignatureBaseString(t *testing.T) {
	got := SignatureBaseString("POST", "https://api.x.com/1.1/statuses/update.json", vectorParams())
	if got != vectorBase {
		t.Fatalf("base string mismatch\n got: %s\nwant: %s", got, vectorBase)
	}
}

func TestSign(t *testing.T) {
	got := vectorSigner().Sign(vectorBase)
	if want := "Ls93hJiZbQ3akF3HF3x1Bz8/zU4="; got != want {
		t.Fatalf("signature = %q, want %q", got, want)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	q := url.Values{"include_entities": {"true"}}
	form := url.Values{"status": {"Hello Ladies + Gentlemen, a signed OAuth request!"}}
	h := vectorSigner().AuthorizationHeader("POST", "https://api.x.com/1.1/statuses/update.json", q, form, nil)
	want := `OAuth oauth_consumer_key="xvz1evFS4wEEPTGEFPHBog", oauth_nonce="kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg", oauth_signature="Ls93hJiZbQ3akF3HF3x1Bz8%2FzU4%3D", oauth_signature_method="HMAC-SHA1", oauth_timestamp="1318622958", oauth_token="370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb", oauth_version="1.0"`
	if h != want {
		t.Fatalf("header mismatch\n got: %s\nwant: %s", h, want)
	}
}

func TestPercentEncode(t *testing.T) {
	if got := PercentEncode("a b+c~-._!*'()/é"); got != "a%20b%2Bc~-._%21%2A%27%28%29%2F%C3%A9" {
		t.Fatalf("got %s", got)
	}
	if strings.Contains(EncodeQuery(url.Values{"a": {"x y"}}), "+") {
		t.Fatal("space must not be +")
	}
}
