package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/open-policy-agent/opa/rego"
)

func main() {
	if err := genToken(); err != nil {
		log.Fatalln(err)
	}
}

func genToken() error {
	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "123456",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)

	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "private"

	// Generate a new private key
	// privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	// if err != nil {
	// 	return fmt.Errorf("generating key: %w", err)
	// }

	file, err := os.Open("zarf/keys/private.pem")
	if err != nil {
		return fmt.Errorf("Open private.pem file: %w", err)
	}
	defer file.Close()

	pemData, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("Reading auth private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return fmt.Errorf("Parsing auth private key: %w", err)
	}

	str, err := token.SignedString(privateKey)

	if err != nil {
		return fmt.Errorf("signed token: %w", err)
	}

	fmt.Println("****** TOKEN ******")
	fmt.Println(str)
	fmt.Println()
	// ==============================================

	fmt.Println("****** PUBLIC KEY ******")

	// Marshal the public key from the private key to PKIX
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the publick key
	publickBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	var b bytes.Buffer
	// Write the private key to the private key file.
	if err := pem.Encode(&b, &publickBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// ============================================================
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))
	var clm struct {
		jwt.RegisteredClaims
		Roles []string
	}

	kf := func(jwt *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}
	tkn, err := parser.ParseWithClaims(str, &clm, kf)
	if err != nil {
		return fmt.Errorf("parsing claims : %w", err)
	}

	if !tkn.Valid {
		return fmt.Errorf("token not valid")
	}

	fmt.Println("TOKEN VALID")

	// ==============================================
	if err := opaPolicyEvaluationAuthen(context.Background(), b.String(), str, clm.Issuer); err != nil {
		return fmt.Errorf("OPS authentication failed: %w", err)
	}
	fmt.Println("TOKEN VALID OPA")
	if err := opaPolicyEvaluationAuthor(context.Background()); err != nil {
		return fmt.Errorf("OPS authorization	 failed: %w", err)
	}
	fmt.Println("AUTH VALIDATED BY OPA")

	fmt.Printf("\n%#v\n", clm)
	return nil
}

func genKey() error {
	// Creating a private key and write it to a file.
	// Generate a new private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}
	// =============================================================
	// Creating a public key and write it to a file
	// Create a file for the private key information in PEM form.
	publickFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publickFile.Close()

	// Marshal the public key from the private key to PKIX
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the publick key
	publickBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the private key to the private key file.
	if err := pem.Encode(publickFile, &publickBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	fmt.Println("private and public key files generated")

	return nil
}

// Core OPA policies.
// Here we are asking the go compiler when you build that file
// I want you to take the contents and embed it in this variable
// This is a way to do it without any file opennings
var (

	//go:embed rego/authentication.rego
	opaAuthentication string

	//go:embed rego/authorization.rego
	opaAuthorization string
)

func opaPolicyEvaluationAuthor(ctx context.Context) error {
	const rule string = "ruleAdminOnly"
	const opaPackage string = "rego"
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)
	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthorization),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}
	input := map[string]any{
		"Roles":   []string{"ADMIN"},
		"Subject": "1234567",
		"UserID":  "1234567",
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}
	return nil
}

// opaPolicyEvaluation asks opa to evaulate the token against the specified token
// policy and public key.
func opaPolicyEvaluationAuthen(ctx context.Context, pem string, tokenString string, issuer string) error {
	const rule string = "auth"
	const opaPackage string = "rego"

	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)
	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthentication),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	input := map[string]any{
		"Key":   pem,
		"Token": tokenString,
		"ISS":   issuer,
	}
	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}
