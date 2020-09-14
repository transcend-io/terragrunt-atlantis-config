FROM golang AS build

ENV GO111MODULE=on
WORKDIR /app

# copy source
COPY go.mod go.sum main.go ./

# fetch deps separately (for layer caching)
RUN go mod download

# build the executable
COPY cmd ./cmd
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# create super thin container with the binary only
FROM scratch
COPY --from=build /app/terragrunt-atlantis-config /app/terragrunt-atlantis-config
ENTRYPOINT [ "/app/terragrunt-atlantis-config" ]
