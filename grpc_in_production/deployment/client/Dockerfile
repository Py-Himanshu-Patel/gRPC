# Multi stage build

# Build stage I : Go lang and Alpine Linux is only needed to build the program
FROM golang AS build

WORKDIR /app

ADD ./client client
ADD ./proto-gen proto-gen
ADD go.mod .

# Download and install all dependencies from go.mod file in /server
RUN go mod tidy
RUN go install ./client

RUN CGO_ENABLED=0 go build -C /app/client -o /bin/grpc-productinfo-client

# Build stage II : Go binaries are self-contained executables.
FROM alpine

# install bash shell in alpine or use default ash shell
RUN apk add --no-cache bash

COPY --from=build /bin/grpc-productinfo-client /bin/grpc-productinfo-client

ENTRYPOINT ["/bin/grpc-productinfo-client"]
EXPOSE 50051