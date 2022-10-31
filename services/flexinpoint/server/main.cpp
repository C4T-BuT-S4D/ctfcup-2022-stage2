#include "server/server.h"
#include <cstdlib>
#include <fmt/format.h>

constexpr size_t kServicePort = 50051;

int main() {
  flexinpoint::FlexinPointService service(std::getenv("PG_DSN"));

  grpc::ServerBuilder builder;
  builder.AddListeningPort(fmt::format("0.0.0.0:{}", kServicePort),
                           grpc::InsecureServerCredentials());
  builder.RegisterService(&service);
  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
  std::cerr << "Let's go" << std::endl;

  server->Wait();
}
