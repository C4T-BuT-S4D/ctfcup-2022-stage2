#pragma once

#include "structs/flexinpoint.grpc.fb.h"
#include "structs/flexinpoint_generated.h"
#include <grpcpp/grpcpp.h>
#include <optional>
#include <pqxx/pqxx>
#include <random>
#include <string>

namespace flexinpoint {

class FlexinPointService final : public FlexinPoint::Service {
public:
  FlexinPointService(std::string connection_string);

  grpc::Status
  Register(grpc::ServerContext *context,
           const flatbuffers::grpc::Message<RegisterRequest> *request,
           flatbuffers::grpc::Message<RegisterResponse> *response) override;

  grpc::Status Me(grpc::ServerContext *context,
                  const flatbuffers::grpc::Message<MeRequest> *request,
                  flatbuffers::grpc::Message<MeResponse> *response) override;

  grpc::Status
  AddStation(grpc::ServerContext *context,
             const flatbuffers::grpc::Message<AddStationRequest> *request,
             flatbuffers::grpc::Message<AddStationResponse> *response) override;

  grpc::Status
  AddRoad(grpc::ServerContext *context,
          const flatbuffers::grpc::Message<AddRoadRequest> *request,
          flatbuffers::grpc::Message<AddRoadResponse> *response) override;

  grpc::Status
  FindPath(grpc::ServerContext *context,
           const flatbuffers::grpc::Message<FindPathRequest> *request,
           flatbuffers::grpc::Message<FindPathResponse> *response) override;

private:
  std::optional<std::string> get_username(std::string_view key);

  pqxx::connection connection_;
  std::random_device random_;
};

} // namespace flexinpoint
