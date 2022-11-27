add_executable(flexinpoint_server main.cpp server.cpp)
target_include_directories(flexinpoint_server PRIVATE ${PROJECT_SOURCE_DIR})
target_link_libraries(flexinpoint_server PRIVATE structs libpqxx::libpqxx fmt::fmt)
