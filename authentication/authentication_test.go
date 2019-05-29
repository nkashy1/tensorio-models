package authentication

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func Test_AtomicAssign(t *testing.T) {
	x := &AuthenticationTokenTypeToSet{"X": {}, "Y": {}}
	y := &AuthenticationTokenTypeToSet{"A": {}}
	atomicAssign(&x, y)
	_, exists := (*x)["A"]
	assert.True(t, exists)
	_, exists = (*x)["X"]
	assert.False(t, exists)
}

func Test_CheckAuthentication(t *testing.T) {
	auth := &GCSAuthentication{
		BucketName:    "",
		TokenFilePath: "/tmp/dummy.txt",
		TokenTypeToSet: &AuthenticationTokenTypeToSet{
			"TokenType1": AuthenticationTokenSet{"Bearer XXXX": {}, "Bearer: YYY": {}},
			"TokenType2": AuthenticationTokenSet{"Bearer ZZZ": {}},
		},
	}

	ctx := context.Background()
	err := auth.CheckAuthentication(ctx, "SomeToken")
	assert.Error(t, err)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"FunnyHeaders:": {"A", "B", "C"},
		"authorization": {"Bearer XXXX"},
	})
	err = auth.CheckAuthentication(ctx, "TokenType1")
	assert.NoError(t, err)
	err = auth.CheckAuthentication(ctx, "TokenType2")
	assert.Error(t, err)
	err = auth.CheckAuthentication(ctx, "TokenType3")
	assert.Error(t, err)
}

func Test_TokenReload(t *testing.T) {
	// This calls Realod under hood (+ cover constructor).
	auth := NewAuthenticator(&FileSystemAuthentication{
		TokenFilePath: "../common/fixtures/AuthTokens.txt",
		TokenTypeToSet: &AuthenticationTokenTypeToSet{
			"TokenType2":  AuthenticationTokenSet{"Bearer ZZZ": {}},
			"FleaClient":  AuthenticationTokenSet{},
			"ModelsAdmin": AuthenticationTokenSet{},
		},
	})
	tokens := *auth.getTokenTypeToSet()
	printTokenSets(tokens)
	assert.Equal(t, len(tokens), 3) // Same token types as in template above.
	empty, exists := tokens["TokenType2"]
	assert.True(t, exists)
	assert.Equal(t, len(empty), 0) // Not in AuthTokens.txt, old value removed.

	// Check that tokens in file authorize.
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"FunnyHeaders:": {"A", "B", "C"},
		"authorization": {"Bearer ClientToken"},
	})
	err := auth.CheckAuthentication(ctx, "FleaClient")
	assert.NoError(t, err)

	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"FunnyHeaders:": {"A", "B", "C"},
		"authorization": {"Bearer ModelsAdminToken"},
	})
	err = auth.CheckAuthentication(ctx, "ModelsAdmin")
	assert.NoError(t, err)
}
