add_executable(flexinpoint_checker main.cpp checker.cpp client.cpp)
target_include_directories(flexinpoint_checker PRIVATE ${PROJECT_SOURCE_DIR})
target_link_libraries(flexinpoint_checker PRIVATE structs fmt::fmt)
