package trace

import (
	"runtime/debug"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	gcodes "google.golang.org/grpc/codes"
	"k8s.io/utils/env"
)

const (
	// GRPCStatusCodeKey is convention for numeric status code of a gRPC request.
	GRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// RPCNameKey is the name of message transmitted or received.
	RPCNameKey = attribute.Key("name")
	// RPCMessageTypeKey is the type of message transmitted or received.
	RPCMessageTypeKey = attribute.Key("message.type")
	// RPCMessageIDKey is the identifier of message transmitted or received.
	RPCMessageIDKey = attribute.Key("message.id")
	// RPCMessageCompressedSizeKey is the compressed size of the message transmitted or received in bytes.
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")
	// RPCMessageUncompressedSizeKey is the uncompressed size of the message
	// transmitted or received in bytes.
	RPCMessageUncompressedSizeKey = attribute.Key("message.uncompressed_size")

	// Name of microservice
	AppName = attribute.Key("app.name")
	// Name of k8s pod
	PodNameKey = attribute.Key("k8s.pod.name")
	// Namespace of k8s pod
	PodNamespaceKey = attribute.Key("k8s.pod.ns")
	// The revision identifier for the current commit or checkout
	VcsRevisionKey = attribute.Key("vcs.rev")
)

// Semantic conventions for common RPC attributes.
var (
	// RPCSystemGRPC is the semantic convention for gRPC as the remoting system.
	RPCSystemGRPC = semconv.RPCSystemKey.String("grpc")
	// RPCNameMessage is the semantic convention for a message named message.
	RPCNameMessage = RPCNameKey.String("message")
	// RPCMessageTypeSent is the semantic conventions for sent RPC message types.
	RPCMessageTypeSent = RPCMessageTypeKey.String("SENT")
	// RPCMessageTypeReceived is the semantic conventions for the received RPC message types.
	RPCMessageTypeReceived = RPCMessageTypeKey.String("RECEIVED")

	PodName      = PodNameKey.String(env.GetString("POD_NAME", "-"))
	PodNamespace = PodNamespaceKey.String(env.GetString("POD_NAMESPACE", "-"))
	VcsRevision  = VcsRevisionKey.String(GetVcsRevision())
)

// StatusCodeAttr returns an attribute.KeyValue that represents the give c.
func StatusCodeAttr(c gcodes.Code) attribute.KeyValue {
	return GRPCStatusCodeKey.Int64(int64(c))
}

func GetVcsRevision() string {
	var rev string
	info, _ := debug.ReadBuildInfo()
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" {
			rev = s.Value
		}
	}
	return rev
}
