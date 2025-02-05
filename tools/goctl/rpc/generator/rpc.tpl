syntax = "proto3";

package {{.package}};{{if .validate}}
import "validate/validate.proto";
{{end}}{{if .gateway}}
import "google/api/annotations.proto";
{{end}}
option go_package="./{{.package}}";

message Request {
  string ping = 1;
}

message Response {
  string pong = 1;
}

service {{.serviceName}} {
  rpc Ping(Request) returns(Response){{if .gateway}}{
    option (google.api.http) = {
      post: "/ping"
      body: "*"
    };
  }{{end}};
}
