# Build stage I : Go lang and Alpine Linux is only needed to build the program

FROM golang AS build

# setting workdir means subsequent directory inside Dockerfile
# will be relative to this directory
WORKDIR /app

# host ./server  to docker image /app/server
ADD ./server server
ADD ./proto-gen proto-gen
ADD go.mod .
# ls -l
#-rw-rw-r-- 1 root root  404 Jun 11 13:57 go.mod
#drwxr-xr-x 3 root root 4096 Jun 10 12:14 proto-gen
#drwxr-xr-x 2 root root 4096 Jun 11 14:29 server

# Download and install all dependencies from go.mod file in /server
RUN go mod tidy
RUN go install ./server

# build go binary and place in /bin/grpc-productinfo-server
RUN CGO_ENABLED=0 go build -C /app/server -o /bin/grpc-productinfo-server

# Build stage II : Go binaries are self-contained executables.
FROM alpine
#FROM scratch - using scratch image doesn't provide shell access

# install bash shell in alpine or use default ash shell
RUN apk add --no-cache bash

# copy from /bin/grpc-productinfo-server in prev build
COPY --from=build /bin/grpc-productinfo-server /bin/grpc-productinfo-server

ENTRYPOINT ["/bin/grpc-productinfo-server"]
EXPOSE 50051