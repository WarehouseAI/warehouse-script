package domain

type CtxKey string

const (
	JsonContentType  = "application/json"
	ProtoContentType = "application/x-protobuf"

	HeaderContentType   = "Content-Type"
	HeaderXForwardedFor = "X-Forwarded-For"

	HeaderAuth       = "Authorization"
	HeaderVersion    = "Coffee-Version"
	VersionDelimiter = ":"

	S3Endpoint = "storage.yandexcloud.net"

	AccountCtxKey     = CtxKey("account")
	TokenNumberCtxKey = CtxKey("token_number")
)
