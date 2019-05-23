package signedURL

import (
	"errors"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

type URLSigningSecrets struct {
	googleAccessID string // client_email in creds.json
	privateKey     []byte // key for the service account in PEM format (generate P12)
	bucketName     string // not really a secret, more for convenience
}

type URLSigner interface {
	GetSignedURL(method string, filePath string, expires time.Time, contentType string) (string, error)
}

// GetSignedURL - creates a signed URL with given expiration, method and content type. Anyone with that URL can
// perform the specified action.
func (secrets URLSigningSecrets) GetSignedURL(method string, filePath string, expires time.Time, contentType string) (string, error) {
	return storage.SignedURL(secrets.bucketName, filePath,
		&storage.SignedURLOptions{
			GoogleAccessID: secrets.googleAccessID,
			PrivateKey:     secrets.privateKey,
			Method:         method,      // e.g. GET, PUT, etc.
			Expires:        expires,     // When access to this would expire.
			ContentType:    contentType, // Important: Must match!
		})
}

// NewURLSignerFromEnvVar - uses GOOGLE_ACCESS_ID and PRIVATE_PEM_KEY environment variables to
// create and return a new URLSigner.
func NewURLSignerFromEnvVar(bucketName string) URLSigner {
	accessID := os.Getenv("GOOGLE_ACCESS_ID") // client_email in creds.json
	if accessID == "" {
		err := errors.New("GOOGLE_ACCESS_ID must be specified")
		panic(err)
	}

	privateKey := os.Getenv("PRIVATE_PEM_KEY") // Generate via P12 creds - see test.
	if privateKey == "" {
		err := errors.New("PRIVATE_PEM_KEY must be specified")
		panic(err)
	}
	signer := URLSigningSecrets{
		googleAccessID: accessID,
		privateKey:     []byte(privateKey),
		bucketName:     bucketName,
	}
	return signer
}
