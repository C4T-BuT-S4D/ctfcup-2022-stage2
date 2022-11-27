#include "checker.h"
#include "checker/client.h"
#include <fmt/format.h>

namespace checker {

namespace {

constexpr size_t kServicePort = 50051;
constexpr size_t kUsernameSize = 32;
constexpr size_t kStationNameSize = 32;
constexpr size_t kStationDescriptionSize = 32;
constexpr size_t kAttributeSize = 12;
constexpr size_t kKeySize = 64;
constexpr std::string_view kAlphabet{"abcdefghijklmnopqrstuvwxyz"};
constexpr std::string_view kName{"_name"};
constexpr std::string_view kDescription{"_description"};

} // namespace

size_t CheckMachine::rnd_number(size_t l, size_t r) {
  if (l >= r) {
    check_failed("bad l, r in rnd_number", "bad l, r in rnd_number");
  }
  return std::uniform_int_distribution<size_t>(l, r - 1)(random_);
}

std::string CheckMachine::rnd_string(size_t length) {
  std::string str;
  str.reserve(length);
  for (size_t i = 0; i < length; ++i) {
    str.push_back(kAlphabet[rnd_number(0, kAlphabet.size())]);
  }
  return str;
}

std::string CheckMachine::rnd_username() { return rnd_string(kUsernameSize); }

std::string CheckMachine::rnd_name() { return rnd_string(kStationNameSize); }

std::string CheckMachine::rnd_description() {
  return rnd_string(kStationDescriptionSize);
}

std::map<std::string, std::string> CheckMachine::rnd_attributes() {
  std::map<std::string, std::string> attributes;
  auto num_attributes = rnd_number(2, 5);
  for (size_t i = 0; i < num_attributes; ++i) {
    auto k = rnd_string(kAttributeSize);
    auto v = rnd_string(kAttributeSize);
    attributes[std::move(k)] = std::move(v);
  }
  return attributes;
}

std::pair<std::string, std::string> CheckMachine::rnd_attribute(
    const std::map<std::string, std::string> &attributes) {
  auto it = attributes.begin();
  auto n = rnd_number(0, attributes.size());
  std::advance(it, n);
  return *it;
}

void CheckMachine::check() {
  auto username = rnd_username();
  auto key = client_.register_(username, mumble);

  std::vector<std::string> names(3);
  std::vector<std::string> descriptions(3);
  std::vector<std::map<std::string, std::string>> attributes(3);
  for (size_t i = 0; i < 3; ++i) {
    names[i] = rnd_name();
    descriptions[i] = rnd_description();
    attributes[i] = rnd_attributes();
    auto station = attributes[i];
    station[std::string(kName)] = names[i];
    station[std::string(kDescription)] = descriptions[i];
    client_.add_station(key, station, mumble);
  }

  client_.add_road(key, names[0], names[1], 4, mumble);
  client_.add_road(key, names[0], names[2], 1, mumble);
  client_.add_road(key, names[1], names[2], 2, mumble);

  std::map<std::string, std::string> filter;
  for (const auto &station_attributes : attributes) {
    auto attribute = rnd_attribute(station_attributes);
    filter[attribute.first] = attribute.second;
  }

  auto path = client_.find_path(key, filter, names[0], names[1], mumble);

  if (path.second != 3) {
    mumble("invalid path length", "invalid path length");
  }

  if (path.first !=
      descriptions[1] + "|" + descriptions[2] + "|" + descriptions[0] + "|") {
    mumble("invalid path string", "invalid path string");
  }

  auto username_ = client_.me(key, mumble);
  if (username != username_) {
    mumble("invalid username on me", "invalid username on me");
  }

  up("OK", "check OK");
}

void CheckMachine::put(std::string_view /*flag_id*/, std::string_view flag,
                       std::string_view /*vuln*/) {
  auto username = rnd_username();
  auto key = client_.register_(username, mumble);

  if (key.size() != kKeySize) {
    mumble("invalid key size", "invalid key size");
  }

  std::vector<std::string> names(3);
  std::vector<std::string> descriptions(3);
  std::vector<std::map<std::string, std::string>> attributes(3);
  for (size_t i = 0; i < 3; ++i) {
    names[i] = rnd_name();
    if (i == 2) {
      descriptions[i] = flag;
    } else {
      descriptions[i] = rnd_description();
    }
    attributes[i] = rnd_attributes();
    auto station = attributes[i];
    station[std::string(kName)] = names[i];
    station[std::string(kDescription)] = descriptions[i];
    client_.add_station(key, station, mumble);
  }

  client_.add_road(key, names[0], names[1], 4, mumble);
  client_.add_road(key, names[0], names[2], 1, mumble);
  client_.add_road(key, names[1], names[2], 2, mumble);

  auto attribute0 = rnd_attribute(attributes[0]);
  auto attribute1 = rnd_attribute(attributes[1]);
  auto attribute2 = rnd_attribute(attributes[2]);

  auto data =
      fmt::format("{}:{}:{}:{}:{}:{}:{}:{}:{}", key, names[0], names[1],
                  attribute0.first, attribute0.second, attribute1.first,
                  attribute1.second, attribute2.first, attribute2.second);
  up(names[2], data);
}

void CheckMachine::get(std::string_view flag_id, std::string_view flag,
                       std::string_view /*vuln*/) {
  auto delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto key = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto start = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto finish = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto attribute0_key = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto attribute0_value = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto attribute1_key = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto attribute1_value = flag_id.substr(0, delim);
  flag_id = flag_id.substr(delim + 1);

  delim = flag_id.find(':');
  if (delim == std::string::npos) {
    check_failed("invalid flag_id", "invalid flag_id");
  }

  auto attribute2_key = flag_id.substr(0, delim);
  auto attribute2_value = flag_id.substr(delim + 1);

  std::map<std::string, std::string> filter;
  filter[std::string(attribute0_key)] = std::move(attribute0_value);
  filter[std::string(attribute1_key)] = std::move(attribute1_value);
  filter[std::string(attribute2_key)] = std::move(attribute2_value);
  auto path = client_.find_path(key, filter, start, finish, corrupt);

  if (path.first.find(flag) == std::string::npos) {
    corrupt("invalid path string", "invalid path string");
  }

  up("OK", "get OK");
}

std::function<void()> action(int argc, const char *argv[]) {
  if (argc < 2) {
    check_failed("", "too few arguments");
  }

  std::string_view method{argv[1]};

  if (method == "info") {
    return []() { up(R"({"vulns":1,"timeout":10,"attack_data":true})", ""); };
  }

  if (argc < 3) {
    check_failed("", "too few arguments");
  }
  auto checker = std::make_shared<CheckMachine>(
      FlexinPointClient{fmt::format("{}:{}", argv[2], kServicePort)});

  if (method == "check") {
    return [checker{std::move(checker)}]() {
      try {
        checker->check();
      } catch (const ServiceUnavailableError &e) {
        down(e.what(), e.what());
      } catch (const ServiceApiError &e) {
        mumble(e.what(), e.what());
      }
    };
  } else if (method == "put") {
    return [argc, argv, checker{std::move(checker)}]() {
      if (argc < 6) {
        check_failed("", "too few arguments");
      }

      try {
        checker->put(argv[3], argv[4], argv[5]);
      } catch (const ServiceUnavailableError &e) {
        down(e.what(), e.what());
      } catch (const ServiceApiError &e) {
        mumble(e.what(), e.what());
      }
    };
  } else if (method == "get") {
    return [argc, argv, checker{std::move(checker)}]() {
      if (argc < 6) {
        check_failed("", "too few arguments");
      }

      try {
        checker->get(argv[3], argv[4], argv[5]);
      } catch (const ServiceUnavailableError &e) {
        down(e.what(), e.what());
      } catch (const ServiceApiError &e) {
        corrupt(e.what(), e.what());
      }
    };
  }

  check_failed("", fmt::format("unknown method: {}", method));
}

} // namespace checker
