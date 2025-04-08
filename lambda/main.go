package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/fujiwara/ridge"
	"github.com/handlename/ssmwrap/v2"
)

type base64secret struct {
	secret []byte
}

func (b *base64secret) UnmarshalText(text []byte) error {
	if b == nil {
		b = &base64secret{}
	}
	decoded, err := base64.RawStdEncoding.DecodeString(string(text))
	if err != nil {
		return err
	}
	b.secret = decoded
	return nil
}

func (b base64secret) Bytes() []byte {
	return b.secret
}

type pemPrivateKey struct {
	key *rsa.PrivateKey
}

func (p *pemPrivateKey) UnmarshalText(text []byte) error {
	if p == nil {
		p = &pemPrivateKey{}
	}

	block, _ := pem.Decode(text)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}
		p.key = key
		return nil
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("not an RSA private key")
		}
		p.key = rsaKey
		return nil
	default:
		return fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}
}

type Options struct {
	ClientID            string         `kong:"name='client-id',help='Client ID',required,env='CLIENT_ID'"`
	ClientSecret        string         `kong:"name='client-secret',help='Client Secret',required,env='CLIENT_SECRET'"`
	SessionSecret       base64secret   `kong:"name='session-secret',help='Session Secret',required,env='SESSION_SECRET'"`
	BaseURL             string         `kong:"name='base-url',help='Base URL',required,env='BASE_URL'"`
	OIDCIssuer          string         `kong:"name='oidc-issuer',help='OIDC Issuer',required,env='OIDC_ISSUER'"`
	AllowedDomains      []string       `kong:"name='allowed-domains',help='Allowed Domains',env='ALLOWED_DOMAINS'"`
	PresignCookieAge    time.Duration  `kong:"name='presign-cookie-age',help='Presign Cookie Age',default='72h',env='PRESIGN_COOKIE_AGE'"`
	RestrictPath        string         `kong:"name='restrict-path',help='Restrict Path',required,env='RESTRICT_PATH'"`
	SignPrivateKey      *pemPrivateKey `kong:"name='sign-private-key',help='Sign Private Key',required,env='SIGN_PRIVATE_KEY'"`
	CloudFrontKeyPairID string         `kong:"name='cloudfront-key-pair-id',help='CloudFront Key Pair ID',required,env='CLOUDFRONT_KEY_PAIR_ID'"`
}

func main() {
	ctx := context.Background()

	paths, ok := os.LookupEnv("SSMWRAP_PATHS")
	if !ok {
		slog.ErrorContext(ctx, "SSMWRAP_PATHS is required")
		os.Exit(1)
	}
	if err := ssmwrap.Export(
		ctx,
		[]ssmwrap.ExportRule{
			{Path: paths},
		},
		ssmwrap.ExportOptions{
			Retries: 3,
		},
	); err != nil {
		slog.ErrorContext(ctx, "failed to export SSM parameters", slog.Any("err", err))
		os.Exit(1)
	}

	var opts Options
	kong.Parse(&opts)

	h, err := newHandler(&opts)
	if err != nil {
		slog.Error("failed to create handler", slog.Any("error", err))
		os.Exit(1)
	}
	ridge.Run(":8080", "/", h)
}
