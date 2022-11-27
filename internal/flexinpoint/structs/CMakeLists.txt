add_library(structs flexinpoint.grpc.fb.cc)
target_link_libraries(structs PUBLIC flatbuffers::flatbuffers gRPC::gRPC)
