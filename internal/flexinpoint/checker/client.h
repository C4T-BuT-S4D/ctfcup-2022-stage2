#pragma once

#include "structs/flexinpoint.grpc.fb.h"
#include "structs/flexinpoint_generated.h"
#include <exception>
#include <fmt/format.h>
#include <grpcpp/grpcpp.h>
#include <grpcpp/support/status.h>

namespace checker {

class ServiceUnavailableError : public std::exception {
public:
  ~ServiceUnavailableError() override = default;

  const char *what() const noexcept override { return "service unavailable"; }
};

class ServiceApiError : public std::exception {
public:
  ServiceApiError(grpc::StatusCode code, std::string error)
      : what_(fmt::format("code {}: {}", code, error)) {}

  ~ServiceApiError() override = default;

  const char *what() const noexcept override { return what_.c_str(); }

private:
  std::string what_;
};

[[noreturn]] inline void handle_error(grpc::Status status) {
  if (status.error_code() == grpc::StatusCode::UNAVAILABLE) {
    throw ServiceUnavailableError();
  }

  throw ServiceApiError(status.error_code(), "unexpected error");
}

using fail_function = std::function<void(std::string, std::string)>;

class FlexinPointClient {
public:
  FlexinPointClient(const std::string &server_address);

  std::string register_(std::string_view username, fail_function fail);
  std::string me(std::string_view key, fail_function fail);
  void add_station(std::string_view key,
                   const std::map<std::string, std::string> &attributes,
                   fail_function fail);
  void add_road(std::string_view key, std::string_view start,
                std::string_view finish, uint32_t length, fail_function fail);
  std::pair<std::string, uint64_t>
  find_path(std::string_view key,
            const std::map<std::string, std::string> &attributes,
            std::string_view start, std::string_view finish,
            fail_function fail);

private:
  std::unique_ptr<flexinpoint::FlexinPoint::Stub> stub_;
};

} // namespace checker
