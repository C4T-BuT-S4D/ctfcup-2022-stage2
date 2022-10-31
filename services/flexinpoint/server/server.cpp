#include "server.h"
#include "structs/flexinpoint_generated.h"
#include <fmt/format.h>
#include <grpcpp/support/status.h>
#include <limits>
#include <optional>
#include <random>

namespace flexinpoint {

namespace {
constexpr size_t kMaxUsernameSize = 64;
constexpr size_t kMaxStationNameSize = 64;
constexpr size_t kKeySize = 64;
constexpr size_t kMaxStationsForPath = 10;
constexpr std::string_view kKeyAlphabet{"0123456789abcdef"};
constexpr std::string_view kName{"_name"};
constexpr std::string_view kDescription{"_description"};

thread_local pqxx::connection connection{std::getenv("PG_DSN")};
} // namespace

FlexinPointService::FlexinPointService() {
  pqxx::work w(connection);

  w.exec(
      fmt::format("CREATE TABLE IF NOT EXISTS users (username VARCHAR({}), key "
                  "CHAR({}))",
                  kMaxUsernameSize, kKeySize));

  w.exec("CREATE TABLE IF NOT EXISTS stations (id SERIAL PRIMARY KEY, "
         "attributes JSONB)");

  w.exec(
      fmt::format("CREATE TABLE IF NOT EXISTS roads (start VARCHAR({}), finish "
                  "VARCHAR({}), length BIGINT, PRIMARY KEY (start, finish))",
                  kMaxStationNameSize, kMaxStationNameSize));

  w.exec(
      fmt::format("CREATE UNIQUE INDEX IF NOT EXISTS stations_idx ON stations "
                  "((attributes ->> {}))",
                  w.quote(kName)));

  w.commit();
}

grpc::Status FlexinPointService::Register(
    grpc::ServerContext * /*context*/,
    const flatbuffers::grpc::Message<RegisterRequest> *request_msg,
    flatbuffers::grpc::Message<RegisterResponse> *response) {
  auto request = request_msg->GetRoot();
  auto username = request->username()->string_view();

  pqxx::work w(connection);

  auto r = w.query01<std::string>(fmt::format(
      "SELECT username FROM users WHERE username = {}", w.quote(username)));

  if (r.has_value()) {
    return {grpc::StatusCode::ALREADY_EXISTS, "user already exists"};
  }

  std::uniform_int_distribution<unsigned char> dist(0, kKeyAlphabet.size() - 1);
  std::string key;
  key.reserve(kKeySize);
  for (size_t i = 0; i < kKeySize; ++i) {
    key.push_back(kKeyAlphabet[dist(random_)]);
  }

  w.exec(fmt::format("INSERT INTO users (username, key) VALUES ({}, {})",
                     w.quote(username), w.quote(key)));
  w.commit();

  flatbuffers::grpc::MessageBuilder mb;
  auto key_offset = mb.CreateString(std::move(key));
  auto response_offset = CreateRegisterResponse(mb, key_offset);
  mb.Finish(response_offset);

  *response = mb.ReleaseMessage<RegisterResponse>();

  return grpc::Status::OK;
}

std::optional<std::string>
FlexinPointService::get_username(std::string_view key) {
  pqxx::work w(connection);

  auto r = w.query01<std::string>(
      fmt::format("SELECT username FROM users WHERE key = {}", w.quote(key)));

  if (!r.has_value()) {
    return {};
  }

  return {std::move(std::get<0>(*r))};
}

grpc::Status
FlexinPointService::Me(grpc::ServerContext * /*context*/,
                       const flatbuffers::grpc::Message<MeRequest> *request_msg,
                       flatbuffers::grpc::Message<MeResponse> *response) {
  auto request = request_msg->GetRoot();

  auto username = get_username(request->key()->string_view());

  if (!username) {
    return {grpc::StatusCode::UNAUTHENTICATED, "user is not authorized"};
  }

  flatbuffers::grpc::MessageBuilder mb;
  auto username_offset = mb.CreateString(std::move(*username));
  auto response_offset = CreateRegisterResponse(mb, username_offset);
  mb.Finish(response_offset);

  *response = mb.ReleaseMessage<MeResponse>();

  return grpc::Status::OK;
}

grpc::Status FlexinPointService::AddStation(
    grpc::ServerContext * /*context*/,
    const flatbuffers::grpc::Message<AddStationRequest> *request_msg,
    flatbuffers::grpc::Message<AddStationResponse> *response) {
  auto request = request_msg->GetRoot();
  auto username = get_username(request->key()->string_view());

  if (!username) {
    return {grpc::StatusCode::UNAUTHENTICATED, "user is not authorized"};
  }

  auto attributes = request->attributes_flexbuffer_root();

  if (!attributes.IsVector()) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "invalid attributes structure"};
  }

  std::string name, description;

  for (size_t i = 0; i < attributes.AsVector().size(); ++i) {
    if (!attributes.AsVector()[i].IsVector() ||
        attributes.AsVector()[i].AsVector().size() != 2 ||
        !attributes.AsVector()[i].AsVector()[0].IsString() ||
        !attributes.AsVector()[i].AsVector()[1].IsString()) {
      return {grpc::StatusCode::INVALID_ARGUMENT,
              "invalid attributes structure"};
    }
    if (attributes.AsVector()[i].AsVector()[0].AsString().str() == kName) {
      name = attributes.AsVector()[i].AsVector()[1].AsString().str();
    } else if (attributes.AsVector()[i].AsVector()[0].AsString().str() ==
               kDescription) {
      description = attributes.AsVector()[i].AsVector()[1].AsString().str();
    }
  }

  if (name.empty()) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "empty name"};
  }

  if (name.size() > kMaxStationNameSize) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "station name too long"};
  }

  if (description.empty()) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "empty description"};
  }

  pqxx::work w(connection);

  auto r = w.query01<uint32_t>(
      fmt::format("SELECT id FROM (SELECT id, value FROM stations s, "
                  "jsonb_array_elements(s.attributes)) AS v WHERE "
                  "v.value ->> 0 = {} AND v.value ->> 1 = {}",
                  w.quote(kName), w.quote(name)));

  if (r.has_value()) {
    return {grpc::StatusCode::ALREADY_EXISTS, "station already exists"};
  }

  w.exec(fmt::format("INSERT INTO stations (attributes) "
                     "VALUES ({})",
                     w.quote(attributes.ToString())));
  w.commit();

  flatbuffers::grpc::MessageBuilder mb;
  auto response_offset = CreateAddStationResponse(mb);
  mb.Finish(response_offset);

  *response = mb.ReleaseMessage<AddStationResponse>();

  return grpc::Status::OK;
}

grpc::Status FlexinPointService::AddRoad(
    grpc::ServerContext * /*context*/,
    const flatbuffers::grpc::Message<AddRoadRequest> *request_msg,
    flatbuffers::grpc::Message<AddRoadResponse> *response) {
  auto request = request_msg->GetRoot();
  auto username = get_username(request->key()->string_view());

  if (!username) {
    return {grpc::StatusCode::UNAUTHENTICATED, "user is not authorized"};
  }

  pqxx::work w(connection);

  auto r = w.query01<uint32_t>(
      fmt::format("SELECT id FROM (SELECT id, value FROM stations s, "
                  "jsonb_array_elements(s.attributes)) AS v WHERE "
                  "v.value ->> 0 = {} AND v.value ->> 1 = {}",
                  w.quote(kName), w.quote(request->start()->string_view())));

  if (!r.has_value()) {
    return {grpc::StatusCode::NOT_FOUND, "start station doesn't exist"};
  }

  r = w.query01<uint32_t>(
      fmt::format("SELECT id FROM (SELECT id, value FROM stations s, "
                  "jsonb_array_elements(s.attributes)) AS v WHERE "
                  "v.value ->> 0 = {} AND v.value ->> 1 = {}",
                  w.quote(kName), w.quote(request->finish()->string_view())));

  if (!r.has_value()) {
    return {grpc::StatusCode::NOT_FOUND, "finish station doesn't exist"};
  }

  auto rr = w.query01<uint32_t>(
      fmt::format("SELECT length FROM roads WHERE start = {} AND finish = {}",
                  w.quote(request->start()->string_view()),
                  w.quote(request->finish()->string_view())));

  if (rr.has_value()) {
    return {grpc::StatusCode::ALREADY_EXISTS, "road already exists"};
  }

  w.exec(fmt::format(
      "INSERT INTO roads (start, finish, length) "
      "VALUES ({}, {}, {}), ({}, {}, {})",
      w.quote(request->start()->string_view()),
      w.quote(request->finish()->string_view()), w.quote(request->length()),
      w.quote(request->finish()->string_view()),
      w.quote(request->start()->string_view()), w.quote(request->length())));
  w.commit();

  flatbuffers::grpc::MessageBuilder mb;
  auto response_offset = CreateAddStationResponse(mb);
  mb.Finish(response_offset);

  *response = mb.ReleaseMessage<AddRoadResponse>();

  return grpc::Status::OK;
}

grpc::Status FlexinPointService::FindPath(
    grpc::ServerContext * /*context*/,
    const flatbuffers::grpc::Message<FindPathRequest> *request_msg,
    flatbuffers::grpc::Message<FindPathResponse> *response) {
  auto request = request_msg->GetRoot();
  auto username = get_username(request->key()->string_view());

  if (!username) {
    return {grpc::StatusCode::UNAUTHENTICATED, "user is not authorized"};
  }

  if (request->start()->string_view() == request->finish()->string_view()) {
    return {grpc::StatusCode::INVALID_ARGUMENT,
            "start and finish stations are the same"};
  }

  pqxx::work w(connection);

  auto r = w.query01<uint32_t>(
      fmt::format("SELECT id FROM (SELECT id, value FROM stations s, "
                  "jsonb_array_elements(s.attributes)) AS v WHERE "
                  "v.value ->> 0 = {} AND v.value ->> 1 = {}",
                  w.quote(kName), w.quote(request->start()->string_view())));

  if (!r.has_value()) {
    return {grpc::StatusCode::NOT_FOUND, "start station doesn't exist"};
  }

  r = w.query01<uint32_t>(
      fmt::format("SELECT id FROM (SELECT id, value FROM stations s, "
                  "jsonb_array_elements(s.attributes)) AS v WHERE "
                  "v.value ->> 0 = {} AND v.value ->> 1 = {}",
                  w.quote(kName), w.quote(request->finish()->string_view())));

  if (!r.has_value()) {
    return {grpc::StatusCode::NOT_FOUND, "finish station doesn't exist"};
  }

  auto attributes = request->attributes_flexbuffer_root();

  if (!attributes.IsVector()) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "invalid attributes structure"};
  }

  if (attributes.AsVector().size() == 0) {
    return {grpc::StatusCode::INVALID_ARGUMENT,
            "filter must contain at least one attribute"};
  }

  std::vector<std::pair<std::string, std::string>> filter;

  for (size_t i = 0; i < attributes.AsVector().size(); ++i) {
    if (!attributes.AsVector()[i].IsVector() ||
        attributes.AsVector()[i].AsVector().size() != 2 ||
        !attributes.AsVector()[i].AsVector()[0].IsString() ||
        !attributes.AsVector()[i].AsVector()[1].IsString()) {
      return {grpc::StatusCode::INVALID_ARGUMENT,
              "invalid attributes structure"};
    }
    if (attributes.AsVector()[i].AsVector()[0].AsString().str() == kName) {
      return {grpc::StatusCode::INVALID_ARGUMENT, "can't search by name"};
    } else if (attributes.AsVector()[i].AsVector()[0].AsString().str() ==
               kDescription) {
      return {grpc::StatusCode::INVALID_ARGUMENT,
              "can't search by description"};
    }
    filter.emplace_back(
        attributes.AsVector()[i].AsVector()[0].AsString().str(),
        attributes.AsVector()[i].AsVector()[1].AsString().str());
  }

  std::string sql =
      fmt::format("SELECT DISTINCT id FROM (SELECT id, value FROM stations s, "
                  "jsonb_array_elements(s.attributes)) as v WHERE v.value ->> "
                  "0 = {} AND v.value ->> 1 = {}",
                  w.quote(filter[0].first), w.quote(filter[0].second));

  for (size_t i = 1; i < filter.size(); ++i) {
    sql += fmt::format(" OR v.value ->> 0 = {} AND v.value ->> 1 = {}",
                       w.quote(filter[i].first), w.quote(filter[i].second));
  }

  auto rr = w.query<uint32_t>(sql);

  std::vector<uint32_t> ids;
  for (auto [id] : rr) {
    ids.push_back(id);
    if (ids.size() > kMaxStationsForPath) {
      return {grpc::StatusCode::INVALID_ARGUMENT,
              "too many stations match filter"};
    }
  }

  if (ids.empty()) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "no stations match filter"};
  }

  sql = fmt::format(
      "SELECT id, v.value ->> 0, v.value ->> 1 FROM (SELECT id, value "
      "FROM stations "
      "s, jsonb_array_elements(s.attributes)) as v WHERE id IN ({}",
      w.quote(ids[0]));

  for (size_t i = 1; i < ids.size(); ++i) {
    sql += fmt::format(", {}", w.quote(ids[i]));
  }

  sql += fmt::format(") AND (v.value ->> 0 = {} OR v.value ->> 0 = {})",
                     w.quote(kName), w.quote(kDescription));

  std::vector<std::tuple<uint32_t, std::string, std::string>> info;
  auto rrr = w.query<uint32_t, std::string, std::string>(sql);
  for (auto &[id, field, value] : rrr) {
    info.emplace_back(id, std::move(field), std::move(value));
    if (info.size() > ids.size() * 2) {
      return {grpc::StatusCode::INVALID_ARGUMENT,
              "too many stations info match filter"};
    }
  }

  if (info.size() != ids.size() * 2) {
    return {grpc::StatusCode::INVALID_ARGUMENT,
            "bad number of stations info match filter"};
  }

  std::map<uint32_t, std::string> names;
  std::map<std::string, std::string> descriptions;

  for (auto &[id, field, value] : info) {
    if (field == kName) {
      names[id] = std::move(value);
    }
  }

  for (uint32_t id : ids) {
    if (!names.contains(id)) {
      return {grpc::StatusCode::INVALID_ARGUMENT, "bad name search info"};
    }
  }

  for (auto &[id, field, value] : info) {
    if (field == kDescription) {
      descriptions[names[id]] = std::move(value);
    }
  }

  if (!descriptions.contains(request->start()->str())) {
    return {grpc::StatusCode::INVALID_ARGUMENT,
            "search doesn't contain start station"};
  }

  if (!descriptions.contains(request->finish()->str())) {
    return {grpc::StatusCode::INVALID_ARGUMENT,
            "search doesn't contain finish station"};
  }

  if (descriptions.empty()) {
    return {grpc::StatusCode::INVALID_ARGUMENT, "empty descriptions search"};
  }

  sql =
      fmt::format("SELECT start, finish, length FROM roads WHERE start IN ({}",
                  w.quote(descriptions.begin()->first));

  auto it = descriptions.begin();
  ++it;
  while (it != descriptions.end()) {
    sql += fmt::format(", {}", w.quote(it->first));
    ++it;
  }

  sql +=
      fmt::format(") AND finish IN ({}", w.quote(descriptions.begin()->first));

  it = descriptions.begin();
  ++it;
  while (it != descriptions.end()) {
    sql += fmt::format(", {}", w.quote(it->first));
    ++it;
  }

  sql += ")";

  auto rrrr = w.query<std::string, std::string, uint32_t>(sql);

  std::map<std::string, std::vector<std::pair<std::string, uint32_t>>> g;

  for (auto &[start, finish, length] : rrrr) {
    g[std::move(start)].emplace_back(std::move(finish), length);
  }

  std::map<std::string, uint64_t> d;
  std::map<std::string, std::optional<std::string>> p;
  for (const auto &[station, _] : descriptions) {
    d[station] = std::numeric_limits<uint64_t>::max();
    p[station] = std::nullopt;
  }

  std::string start(request->start()->str());

  d[start] = 0;
  std::set<std::pair<uint64_t, std::string>> q;
  q.insert({0, start});

  while (!q.empty()) {
    auto current = std::move(q.begin()->second);
    q.erase(q.begin());

    for (auto to : g[current]) {
      if (d[current] + to.second < d[to.first]) {
        q.erase({d[to.first], to.first});
        d[to.first] = d[current] + to.second;
        p[to.first] = current;
        q.insert({d[to.first], to.first});
      }
    }
  }

  std::string path;
  std::optional<std::string> finish(request->finish()->str());
  uint64_t dist = d[*finish];

  while (finish.has_value()) {
    path += descriptions[*finish] + "|";
    finish = p[*finish];
  }

  flatbuffers::grpc::MessageBuilder mb;
  auto response_offset =
      CreateFindPathResponse(mb, mb.CreateString(path), dist);
  mb.Finish(response_offset);

  *response = mb.ReleaseMessage<FindPathResponse>();

  return grpc::Status::OK;
}

} // namespace flexinpoint
