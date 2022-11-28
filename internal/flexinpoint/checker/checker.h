#pragma once

#include "checker/client.h"
#include <exception>
#include <functional>
#include <random>
#include <string>

namespace checker {

enum class Status {
  UP = 101,
  CORRUPT = 102,
  MUMBLE = 103,
  DOWN = 104,
  CHECK_FAILED = 105,
};

class StatusException : public std::exception {
public:
  StatusException(Status status, std::string public_message,
                  std::string private_message)
      : status_(status), public_message_(std::move(public_message)),
        private_message_(std::move(private_message)) {}
  ~StatusException() override = default;

  const char *what() const noexcept override { return "status"; }

  Status status() const { return status_; }

  const std::string &public_message() const { return public_message_; }

  const std::string &private_message() const { return private_message_; }

private:
  Status status_;
  std::string public_message_;
  std::string private_message_;
};

[[noreturn]] inline void up(std::string public_message,
                            std::string private_message) {
  throw StatusException(Status::UP, std::move(public_message),
                        std::move(private_message));
}

[[noreturn]] inline void corrupt(std::string public_message,
                                 std::string private_message) {
  throw StatusException(Status::CORRUPT, std::move(public_message),
                        std::move(private_message));
}

[[noreturn]] inline void mumble(std::string public_message,
                                std::string private_message) {
  throw StatusException(Status::MUMBLE, std::move(public_message),
                        std::move(private_message));
}

[[noreturn]] inline void down(std::string public_message,
                              std::string private_message) {
  throw StatusException(Status::DOWN, std::move(public_message),
                        std::move(private_message));
}

[[noreturn]] inline void check_failed(std::string public_message,
                                      std::string private_message) {
  throw StatusException(Status::CHECK_FAILED, std::move(public_message),
                        std::move(private_message));
}

class CheckMachine {
public:
  CheckMachine(FlexinPointClient client) : client_(std::move(client)) {}

  void check();
  void put(std::string_view flag_id, std::string_view flag,
           std::string_view vuln);
  void get(std::string_view flag_id, std::string_view flag,
           std::string_view vuln);

private:
  size_t rnd_number(size_t l, size_t r);
  std::string rnd_string(size_t length);
  std::string rnd_username();
  std::string rnd_name();
  std::string rnd_description();
  std::map<std::string, std::string> rnd_attributes();
  std::pair<std::string, std::string>
  rnd_attribute(const std::map<std::string, std::string> &attributes);

  FlexinPointClient client_;
  std::random_device random_;
};

std::function<void()> action(int argc, const char *argv[]);

} // namespace checker
