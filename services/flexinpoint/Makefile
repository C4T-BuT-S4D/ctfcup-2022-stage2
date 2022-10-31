.PHONY: all
all: binaries

structs/flexinpoint_generated.h structs/flexinpoint.grpc.fb.cc structs/flexinpoint.grpc.fb.h: structs/flexinpoint.fbs
	rm -f ./structs/*generated.h ./structs/*fb.cc ./structs/*fb.h
	./flatbuffers/flatc --cpp --gen-object-api --scoped-enums --gen-mutable --grpc -o structs structs/flexinpoint.fbs

.PHONY: binaries
binaries: structs/flexinpoint_generated.h structs/flexinpoint.grpc.fb.cc structs/flexinpoint.grpc.fb.h
	cmake -B build
	cd build && make
