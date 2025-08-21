package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"time"

	"substack-auth/pkg/config"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

type JWT struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	expiry     time.Duration
}

func New(cfg *config.Config) (*JWT, error) {
	privateKey, err := loadPrivateKey(cfg.JWT.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey, err := loadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	slog.Info("JWT service initialized", "expiry", cfg.JWT.Expiration)

	return &JWT{
		privateKey: privateKey,
		publicKey:  publicKey,
		expiry:     cfg.JWT.Expiration,
	}, nil
}

func (j *JWT) GenerateToken(username string) (string, error) {
	now := time.Now()
	claims := jwt.Claims{
		Subject:  username,
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(j.expiry)),
	}

	// Create signer
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: j.privateKey}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	token, err := jwt.Signed(signer).Claims(claims).Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return token, nil
}

func (j *JWT) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseSigned(tokenString, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	var claims jwt.Claims
	if err := token.Claims(j.publicKey, &claims); err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}

	if err := claims.Validate(jwt.Expected{Time: time.Now()}); err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}

	return claims.Subject, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	slog.Debug("Attempting to parse private key", "type", block.Type, "size", len(block.Bytes))

	// Try PKCS8 first (more common), then PKCS1
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		slog.Debug("PKCS8 parsing failed, trying PKCS1", "error", err)
		// Try PKCS1 format
		privateKey, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key (tried PKCS8 and PKCS1): %w, %w", err, err2)
		}
		return privateKey, nil
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	return rsaKey, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPublicKey, nil
}
