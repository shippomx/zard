package httpx

import "github.com/shippomx/zard/rest/internal/header"

const (
	// ContentEncoding means Content-Encoding.
	ContentEncoding = "Content-Encoding"
	// ContentSecurity means X-Content-Security.
	ContentSecurity = "X-Content-Security"
	// ContentType means Content-Type.
	ContentType = header.ContentType
	// JsonContentType means application/json.
	JsonContentType = header.JsonContentType
	// KeyField means key.
	KeyField = "key"
	// SecretField means secret.
	SecretField = "secret"
	// TypeField means type.
	TypeField = "type"
	// CryptionType means cryption.
	CryptionType = 1
	// KeyUserId means key user id.
	// nolint:revive // same style as above
	KeyUserId = "X-Gate-User-Id"
)

const (
	// CodeSignaturePass means signature verification passed.
	CodeSignaturePass = iota
	// CodeSignatureInvalidHeader means invalid header in signature.
	CodeSignatureInvalidHeader
	// CodeSignatureWrongTime means wrong timestamp in signature.
	CodeSignatureWrongTime
	// CodeSignatureInvalidToken means invalid token in signature.
	CodeSignatureInvalidToken
)
