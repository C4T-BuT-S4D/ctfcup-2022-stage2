# FROM ubuntu:focal-20221019
# ARG DEBIAN_FRONTEND=noninteractive
# ENV TZ=Europe/Moscow

# RUN apt -y update && apt -y install git cmake make g++ python3 python3-pip
# RUN python3 -m pip install conan

# WORKDIR /app
# RUN git clone https://github.com/google/flatbuffers.git

# WORKDIR /app/flatbuffers
# RUN git checkout v2.0.8 && cmake -G "Unix Makefiles" -DCMAKE_BUILD_TYPE=Release && make -j16

# WORKDIR /app
# COPY CMakeLists.txt .
# COPY conan conan
# RUN cmake -B build

# Pre-built dependencies, no backdoors
FROM ae64f3efc13cf0cb/flexinpoint_base:latest
WORKDIR /app
COPY . .
RUN make -j4

FROM ubuntu:focal-20221019
COPY --from=0 /app/build/server/flexinpoint_server /
CMD ["/flexinpoint_server"]
