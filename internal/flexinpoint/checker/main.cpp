#include "checker/checker.h"
#include <exception>
#include <iostream>

int main(int argc, const char *argv[]) {
  try {
    auto action = checker::action(argc, argv);
    action();
  } catch (const checker::StatusException &e) {
    std::cout << e.public_message() << std::endl;
    std::cerr << e.private_message() << std::endl;
    return static_cast<int>(e.status());
  } catch (const std::exception &e) {
    std::cerr << e.what() << std::endl;
    return static_cast<int>(checker::Status::CHECK_FAILED);
  }
}
