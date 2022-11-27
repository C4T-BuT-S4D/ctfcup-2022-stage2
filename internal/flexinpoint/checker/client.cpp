#include "client.h"
#include "checker/checker.h"
#include "structs/flexinpoint_generated.h"
#include <grpcpp/support/status.h>

namespace checker {

FlexinPointClient::FlexinPointClient(const std::string &server_address)
    : stub_(flexinpoint::FlexinPoint::NewStub(grpc::CreateChannel(
          server_address, grpc::InsecureChannelCredentials()))) {}

std::string FlexinPointClient::register_(std::string_view username,
                                         fail_function fail) {
  flatbuffers::grpc::MessageBuilder mb;

  auto request_offset =
      flexinpoint::CreateRegisterRequest(mb, mb.CreateString(username));
  mb.Finish(request_offset);
  auto request = mb.ReleaseMessage<flexinpoint::RegisterRequest>();

  flatbuffers::grpc::Message<flexinpoint::RegisterResponse> response_msg;
  grpc::ClientContext context;

  auto status = stub_->Register(&context, request, &response_msg);

  if (status.ok()) {
    return response_msg.GetRoot()->key()->str();
  }

  if (status.error_code() == grpc::StatusCode::ALREADY_EXISTS) {
    fail("can't register", "can't register, already exists");
  }

  handle_error(status);
}

std::string FlexinPointClient::me(std::string_view key, fail_function fail) {
  flatbuffers::grpc::MessageBuilder mb;

  auto request_offset = flexinpoint::CreateMeRequest(mb, mb.CreateString(key));
  mb.Finish(request_offset);
  auto request = mb.ReleaseMessage<flexinpoint::MeRequest>();

  flatbuffers::grpc::Message<flexinpoint::MeResponse> response_msg;
  grpc::ClientContext context;

  auto status = stub_->Me(&context, request, &response_msg);

  if (status.ok()) {
    return response_msg.GetRoot()->username()->str();
  }

  if (status.error_code() == grpc::StatusCode::UNAUTHENTICATED) {
    fail("can't get me", "can't get me, unauthorized");
  }

  handle_error(status);
}

void FlexinPointClient::add_station(
    std::string_view key, const std::map<std::string, std::string> &attributes,
    fail_function fail) {
  flexbuffers::Builder fbb;
  fbb.Vector([&fbb, &attributes]() {
    for (const auto &it : attributes) {
      const auto &k = it.first;
      const auto &v = it.second;
      fbb.Vector([&fbb, &k, &v]() {
        fbb.String(k);
        fbb.String(v);
      });
    }
  });
  fbb.Finish();

  flatbuffers::grpc::MessageBuilder mb;

  auto request_offset = flexinpoint::CreateAddStationRequest(
      mb, mb.CreateString(key), mb.CreateVector(fbb.GetBuffer()));
  mb.Finish(request_offset);
  auto request = mb.ReleaseMessage<flexinpoint::AddStationRequest>();

  flatbuffers::grpc::Message<flexinpoint::AddStationResponse> response_msg;
  grpc::ClientContext context;

  auto status = stub_->AddStation(&context, request, &response_msg);

  if (status.ok()) {
    return;
  }

  if (status.error_code() == grpc::StatusCode::UNAUTHENTICATED) {
    fail("can't add station", "can't add station, unauthorized");
  }

  if (status.error_code() == grpc::StatusCode::INVALID_ARGUMENT) {
    fail("can't add station", "can't add station, invalid argument");
  }

  if (status.error_code() == grpc::StatusCode::ALREADY_EXISTS) {
    fail("can't add station", "can't add station, already exists");
  }

  handle_error(status);
}

void FlexinPointClient::add_road(std::string_view key, std::string_view start,
                                 std::string_view finish, uint32_t length,
                                 fail_function fail) {
  flatbuffers::grpc::MessageBuilder mb;

  auto request_offset = flexinpoint::CreateAddRoadRequest(
      mb, mb.CreateString(key), mb.CreateString(start), mb.CreateString(finish),
      length);
  mb.Finish(request_offset);
  auto request = mb.ReleaseMessage<flexinpoint::AddRoadRequest>();

  flatbuffers::grpc::Message<flexinpoint::AddRoadResponse> response_msg;
  grpc::ClientContext context;

  auto status = stub_->AddRoad(&context, request, &response_msg);

  if (status.ok()) {
    return;
  }

  if (status.error_code() == grpc::StatusCode::UNAUTHENTICATED) {
    fail("can't add road", "can't add road, unauthorized");
  }

  if (status.error_code() == grpc::StatusCode::NOT_FOUND) {
    fail("can't add road", "can't add road, not found");
  }

  if (status.error_code() == grpc::StatusCode::ALREADY_EXISTS) {
    fail("can't add road", "can't add road, already exists");
  }

  handle_error(status);
}

std::pair<std::string, uint64_t> FlexinPointClient::find_path(
    std::string_view key, const std::map<std::string, std::string> &attributes,
    std::string_view start, std::string_view finish, fail_function fail) {
  flexbuffers::Builder fbb;
  fbb.Vector([&fbb, &attributes]() {
    for (const auto &it : attributes) {
      const auto &k = it.first;
      const auto &v = it.second;
      fbb.Vector([&fbb, &k, &v]() {
        fbb.String(k);
        fbb.String(v);
      });
    }
  });
  fbb.Finish();

  flatbuffers::grpc::MessageBuilder mb;

  auto request_offset = flexinpoint::CreateFindPathRequest(
      mb, mb.CreateString(key), mb.CreateVector(fbb.GetBuffer()),
      mb.CreateString(start), mb.CreateString(finish));
  mb.Finish(request_offset);
  auto request = mb.ReleaseMessage<flexinpoint::FindPathRequest>();

  flatbuffers::grpc::Message<flexinpoint::FindPathResponse> response_msg;
  grpc::ClientContext context;

  auto status = stub_->FindPath(&context, request, &response_msg);

  if (status.ok()) {
    return {response_msg.GetRoot()->path()->str(),
            response_msg.GetRoot()->length()};
  }

  if (status.error_code() == grpc::StatusCode::UNAUTHENTICATED) {
    fail("can't add station", "can't add station, unauthorized");
  }

  if (status.error_code() == grpc::StatusCode::INVALID_ARGUMENT) {
    fail("can't add station", "can't add station, invalid argument");
  }

  if (status.error_code() == grpc::StatusCode::NOT_FOUND) {
    fail("can't add station", "can't add station, not found");
  }

  handle_error(status);
}

} // namespace checker
