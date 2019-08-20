protoc -I=. --micro_out=. --go_out=. ./proto.proto
move /Y .\github.com\paysuper\paysuper-reporter\pkg\proto.micro.go .\proto.micro.go
move /Y .\github.com\paysuper\paysuper-reporter\pkg\proto.pb.go .\proto.pb.go
rmdir /Q/S .\github.com
protoc-go-inject-tag -input=./proto.pb.go -XXX_skip=bson,json,structure,validate
