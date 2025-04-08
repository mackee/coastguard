package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	CookieNamePolicy    = "CloudFront-Policy"
	CookieNameSignature = "CloudFront-Signature"
	CookieNameKeyPairID = "CloudFront-Key-Pair-Id"
)

type CloudFrontPolicy struct {
	Statement []CloudFrontPolicyStatement `json:"Statement"`
}

type CloudFrontPolicyStatement struct {
	Resource  string                             `json:"Resource"`
	Condition CloudFrontPolicyStatementCondition `json:"Condition"`
}

type CloudFrontPolicyStatementCondition struct {
	DateLessThan CloudFrontPolicyStatementConditionDate `json:"DateLessThan"`
}

type CloudFrontPolicyStatementConditionDate struct {
	EpochTime int64 `json:"AWS:EpochTime"`
}

var cloudFrontPolicyBase64Encoder = base64.
	NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-~").
	WithPadding('_')

func SetCookiesPresign(w http.ResponseWriter, opts *Options) (err error) {
	expires := time.Now().Add(opts.PresignCookieAge)
	resource, err := url.JoinPath(opts.BaseURL, opts.RestrictPath, "*")
	if err != nil {
		return fmt.Errorf("failed to join URL: %w", err)
	}
	policy := CloudFrontPolicy{
		Statement: []CloudFrontPolicyStatement{
			{
				Resource: resource,
				Condition: CloudFrontPolicyStatementCondition{
					DateLessThan: CloudFrontPolicyStatementConditionDate{
						EpochTime: expires.Unix(),
					},
				},
			},
		},
	}
	j, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}
	encoded := cloudFrontPolicyBase64Encoder.EncodeToString(j)
	policyCookie := http.Cookie{
		Name:     CookieNamePolicy,
		Value:    encoded,
		Path:     "/",
		Expires:  expires,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	w.Header().Add("Set-Cookie", policyCookie.String())

	hash := sha1.Sum(j)
	rand := rand.Reader
	signed, err := rsa.SignPKCS1v15(rand, opts.SignPrivateKey.key, crypto.SHA1, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign policy: %w", err)
	}
	encodedSign := cloudFrontPolicyBase64Encoder.EncodeToString(signed)
	signCookie := http.Cookie{
		Name:     CookieNameSignature,
		Value:    encodedSign,
		Path:     "/",
		Expires:  expires,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	w.Header().Add("Set-Cookie", signCookie.String())
	keyPairCookie := http.Cookie{
		Name:     CookieNameKeyPairID,
		Value:    opts.CloudFrontKeyPairID,
		Path:     "/",
		Expires:  expires,
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	w.Header().Add("Set-Cookie", keyPairCookie.String())

	return nil
}

func RemovePresignedCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNamePolicy,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Value:    "",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameSignature,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Value:    "",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieNameKeyPairID,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Value:    "",
	})
}
