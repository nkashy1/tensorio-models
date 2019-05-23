package signedURL

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// To run this test in verbose mode:
//   go clean -testcache ./...
//   go test github.com/doc-ai/tensorio-models/signed_url -test.v -count=1
// Need to set GOOGLE_ACCESS_ID to the client e-mails for the repo.
// Need to set PRIVATE_PEM_KEY to the contents of the *.pem file.
// Need to set FLEA_CGS_UPLOAD_BUCKET to a test bucket name
// To generate a *.pem file do:
//   - Go to ttps://console.cloud.google.com/iam-admin/
//   - Click on the appropriate service account
//   - Edit it and Create a key in legacy P12 format - and downlosd it
//   - Run:  openssl pkcs12 -in key.p12 -passin pass:notasecret -out key.pem -nodes (or whatever password is used)

func Test_URLSigning(t *testing.T) {
	if os.Getenv("GOOGLE_ACCESS_ID") == "" ||
		os.Getenv("PRIVATE_PEM_KEY") == "" {
		fmt.Println("Running this test requires GOOGLE_ACCESS_ID and PRIVATE_PEM_KEY env variables")
		return
	}
	bucket := os.Getenv("FLEA_GCS_UPLOAD_BUCKET")
	if bucket == "" {
		panic("Please, specify FLEA_GCS_UPLOAD_BUCKET to run this test")
	}
	fmt.Println("These URLs will expire in 3 minutes. Please, copy and type the curl commands to test!")
	signer := NewURLSignerFromEnvVar(bucket)
	url, err := signer.GetSignedURL("PUT", "url-signing-test.txt", time.Now().Add(time.Minute*3), "text/plain")
	assert.NoError(t, err)
	fmt.Println("curl -X PUT -H \"Content-Type: text/plain\" -d \"Uploaded at: $(date)\\n\" \"" + url + "\"")
	url, err = signer.GetSignedURL("GET", "url-signing-test.txt", time.Now().Add(time.Minute*3), "") //  "text/plain")
	assert.NoError(t, err)
	fmt.Println("curl \"" + url + "\"")
}
