cmake_minimum_required(VERSION 3.16)
project(flexinpoint LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
add_compile_options(-Wall -Wextra -pedantic -Werror)

include(${PROJECT_SOURCE_DIR}/conan/cmake)

set(CONAN_WORK_DIR ${CMAKE_BINARY_DIR}/conan)
file(MAKE_DIRECTORY ${CONAN_WORK_DIR})

conan_cmake_install(
  PATH_OR_REFERENCE
  ${PROJECT_SOURCE_DIR}/conan/deps
  PROFILE
  ${PROJECT_SOURCE_DIR}/conan/profile
  BUILD
  missing
  INSTALL_FOLDER
  ${CONAN_WORK_DIR})

include(${CONAN_WORK_DIR}/conan_paths.cmake)

find_package(flatbuffers REQUIRED)
find_package(gRPC REQUIRED)
find_package(libpqxx REQUIRED)
find_package(fmt REQUIRED)

if(EXISTS structs)
add_subdirectory(structs)
endif()

if(EXISTS server)
add_subdirectory(server)
endif()

if(EXISTS checker)
add_subdirectory(checker)
endif()

if(EXISTS exploit)
add_subdirectory(exploit)
endif()
