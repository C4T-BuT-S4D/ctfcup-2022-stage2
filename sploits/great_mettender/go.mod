module gmtsploit

go 1.18

require (
	gmtchecker v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.50.1
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace (
	gmtchecker => ../../checkers/great_mettender
	gmtservice => ../../services/great_mettender
)
