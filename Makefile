.PHONY: http_codegen
http_codegen:
	oapi-codegen -generate types -o internal/trainer/ports/openapi_types.gen.go -package ports api/openapi/trainer.yml
	oapi-codegen -generate chi-server -o internal/trainer/ports/openapi_api.gen.go -package ports api/openapi/trainer.yml

.PHONY: proto_codegen
proto_codegen:
	protoc --go_out=internal/common/genproto/trainer --go_opt=paths=source_relative \
	       --go-grpc_out=internal/common/genproto/trainer --go-grpc_opt=paths=source_relative \
	       -I api/protobuf api/protobuf/trainer.proto

	protoc --go_out=internal/common/genproto/users --go_opt=paths=source_relative \
	       --go-grpc_out=internal/common/genproto/users --go-grpc_opt=paths=source_relative \
	       -I api/protobuf api/protobuf/users.proto
# proto_codegen:
# 	protoc --go_out=plugins=grpc:internal/common/genproto/trainer -I api/protobuf api/protobuf/trainer.proto
# 	protoc --go_out=plugins=grpc:internal/common/genproto/users -I api/protobuf api/protobuf/users.proto