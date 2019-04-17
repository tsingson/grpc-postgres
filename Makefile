install:
	go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/jteeuwen/go-bindata/go-bindata \
		github.com/golang/mock/mockgen

generate:
	protoc -I proto -I=$HOME/go/src -I=/usr/local/include --go_out=plugins=grpc,paths=source_relative:./proto ./proto/users.proto
	go-bindata -pkg migrations -ignore bindata -prefix ./users/migrations/ -o ./users/migrations/bindata.go ./users/migrations
	mockgen -destination ./users/mocks_test.go -package users_test github.com/tsingson/grpc-postgres/proto UserService_ListUsersServer
