FROM golang:latest
WORKDIR /build
COPY . ./
RUN apt-get update
RUN apt-get install libsdl2-dev --quiet -y
RUN git config --global http.sslverify false
RUN make
