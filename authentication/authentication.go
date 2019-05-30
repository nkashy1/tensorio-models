package authentication

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"sync/atomic"
	"unsafe"

	gcs "cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Basically a set of valid tokens. Note that the tokens have "Bearer " prepended to them for faster checking.
type AuthenticationTokenSet map[string]struct{}
type AuthenticationTokenType string
type AuthenticationTokenTypeToSet map[AuthenticationTokenType]AuthenticationTokenSet

const NoAuthentication AuthenticationTokenType = "ALLOW-ALL"

type FullMethodName string
type MethodToAuthenticationTokenType map[FullMethodName]AuthenticationTokenType

type GCSAuthentication struct {
	BucketName     string
	TokenFilePath  string
	TokenTypeToSet *AuthenticationTokenTypeToSet
}

type FileSystemAuthentication struct {
	TokenFilePath  string
	TokenTypeToSet *AuthenticationTokenTypeToSet
}

type Authenticator interface {
	ReloadAuthenticationTokens(ctx context.Context) error
	CheckAuthentication(ctx context.Context, tokenType AuthenticationTokenType) error

	getTokenTypeToSet() *AuthenticationTokenTypeToSet // for testing only
}

var (
	ErrMissingHeaders             = errors.New("Unauthorized. Missing headers")
	ErrMissingAuthorizationHeader = errors.New("Unauthorized. Missing Authorization header")
	ErrInvalidAuthorizationToken  = errors.New("Unauthorized. Invalid Authorization token")
	ErrMalformedTokenFile         = errors.New("Malformed token file")
	ErrNoTokensOfSpecifiedType    = errors.New("Unauthorized. No tokens configured")
	ErrNooneIsAuthorized          = errors.New("Unauthorized. Noone is authorized")
)

// LogMethodNameServerInterceptor - do nothing interceptor that just prints the nmethod names for debugging.
// Can be used instead of the output of CreateGRPCServerInterceptor.
func LogMethodNameServerInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	log.Println(info.FullMethod, req)
	return handler(ctx, req)
}

// CreateGRPCInterceptor - given an authenticator and a mapping between methods and token types,
// creates a GRPC middleware interceptor.
// To enable authentication just pass the output or this function to grpc.NewServer() like so:
//   grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))
func CreateGRPCInterceptor(authenticator Authenticator,
	methodToTokenType MethodToAuthenticationTokenType) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		tokenType, exists := methodToTokenType[FullMethodName(info.FullMethod)]
		if !exists {
			// Returns error code 401.
			return nil, status.Errorf(codes.Unauthenticated, "Unauthorized. No one is authorized")
		}
		if tokenType == NoAuthentication {
			log.Println(info.FullMethod, req)
			return handler(ctx, req)
		}
		err := authenticator.CheckAuthentication(ctx, tokenType)
		if err != nil {
			// Error code 401
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}
		log.Println(info.FullMethod, req)
		return handler(ctx, req)
	}
}

// NewAuthenticator - returns an authenticator object. All valid keys in the TokenTypeToSet must be
// specified in template. Panics on failure to load objects.
func NewAuthenticator(template Authenticator) Authenticator {
	err := template.ReloadAuthenticationTokens(context.Background())
	if err != nil {
		panic(err)
	}
	return template
}

func atomicAssign(destSet **AuthenticationTokenTypeToSet, newSet *AuthenticationTokenTypeToSet) {
	atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(destSet)), unsafe.Pointer(newSet))
}

func createEmptyCopyTokenTypeToSet(origSet *AuthenticationTokenTypeToSet) *AuthenticationTokenTypeToSet {
	result := make(AuthenticationTokenTypeToSet)
	for k := range *origSet {
		result[k] = AuthenticationTokenSet{}
	}
	return &result
}

// ParseTokenSetsFile - supports file containing multiple types of tokens and parses the subset we casre about.
// The format of the file is <token-type> <auth-tokemn>
func ParseTokenSetsFile(file io.Reader, tokenTypeToSet *AuthenticationTokenTypeToSet) error {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)
	skippedTokens := 0
	keptTokens := 0
	var dummy struct{}
	for scanner.Scan() {
		tokenType := scanner.Text()
		tokenSet, exists := (*tokenTypeToSet)[AuthenticationTokenType(tokenType)]
		if !scanner.Scan() {
			return ErrMalformedTokenFile
		}
		if !exists {
			skippedTokens++
			continue // We don't care about this token type
		}
		token := "Bearer " + scanner.Text()
		tokenSet[token] = dummy // Insert token in appropriate set
		keptTokens++
	}
	log.Printf("Successfully parsed authentication tokens. #Kept: %d #Skipped: %d",
		keptTokens, skippedTokens)
	return nil
}

func printTokenSets(sets AuthenticationTokenTypeToSet) {
	for k, v := range sets {
		log.Println("TokenType: " + k)
		for t := range v {
			log.Println("Token: " + t)
		}
	}
}

// ReloadAuthenticationTokens - parses a new copy of the authentication tokens and
// thread-safely replaces them
func (auth *GCSAuthentication) ReloadAuthenticationTokens(ctx context.Context) error {
	newSet := createEmptyCopyTokenTypeToSet(auth.TokenTypeToSet)

	client, err := gcs.NewClient(ctx)
	if err != nil {
		return err
	}
	tokenFile := client.Bucket(auth.BucketName).Object(auth.TokenFilePath)
	_, err = tokenFile.Attrs(ctx)
	if err != nil {
		return err
	}
	reader, err := tokenFile.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()
	err = ParseTokenSetsFile(reader, newSet)
	if err != nil {
		return err
	}

	atomicAssign(&auth.TokenTypeToSet, newSet)
	for k, v := range *auth.TokenTypeToSet {
		log.Printf("We have %d auth tokens of type: %s", len(v), k)
	}
	return nil
}

func (auth *FileSystemAuthentication) ReloadAuthenticationTokens(ctx context.Context) error {
	newSet := createEmptyCopyTokenTypeToSet(auth.TokenTypeToSet)
	file, err := os.Open(auth.TokenFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	err = ParseTokenSetsFile(file, newSet)
	if err != nil {
		return err
	}

	atomicAssign(&auth.TokenTypeToSet, newSet)
	for k, v := range *auth.TokenTypeToSet {
		log.Printf("We have %d auth tokens of type: %s", len(v), k)
	}
	return nil
}

func checkAuthentication(ctx context.Context, tokenType AuthenticationTokenType, tokenTypeToSet *AuthenticationTokenTypeToSet) error {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ErrMissingHeaders
	}
	// Seems that grpc lowercases the header-names since they are case-insensitive
	// Also, note that no trailing : here.
	auth_tokens, exists := headers["authorization"]
	if !exists || len(auth_tokens) < 1 {
		return ErrMissingAuthorizationHeader
	}
	tokenSet, exists := (*tokenTypeToSet)[tokenType]
	if !exists {
		return ErrNoTokensOfSpecifiedType
	}
	// Token must include 'Bearer ' prefix.
	_, valid := tokenSet[auth_tokens[0]]
	if !valid {
		return ErrInvalidAuthorizationToken
	}
	return nil // Request is authorized.
}

func (auth *GCSAuthentication) CheckAuthentication(ctx context.Context, tokenType AuthenticationTokenType) error {
	return checkAuthentication(ctx, tokenType, auth.TokenTypeToSet)
}

func (auth *FileSystemAuthentication) CheckAuthentication(ctx context.Context, tokenType AuthenticationTokenType) error {
	return checkAuthentication(ctx, tokenType, auth.TokenTypeToSet)
}

func (auth *GCSAuthentication) getTokenTypeToSet() *AuthenticationTokenTypeToSet {
	return auth.TokenTypeToSet
}

func (auth *FileSystemAuthentication) getTokenTypeToSet() *AuthenticationTokenTypeToSet {
	return auth.TokenTypeToSet
}

////////////////////////////////////////////////////////////////////////

type FakeAuthentication struct{}

func (f FakeAuthentication) ReloadAuthenticationTokens(ctx context.Context) error {
	return nil
}
func (f FakeAuthentication) CheckAuthentication(ctx context.Context, tokenType AuthenticationTokenType) error {
	return nil
}
func (f FakeAuthentication) getTokenTypeToSet() *AuthenticationTokenTypeToSet {
	return nil
}

// Returns a do nothing, authorize-all authenticator for testing.
func NewFakeAuthenticator() Authenticator {
	return FakeAuthentication{}
}
