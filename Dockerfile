# Build the manager binary
FROM golang:1.10.3 as builder

    # Copy in the go src
    WORKDIR /go/src/github.com/subravi92/burrow-operator
    COPY pkg/    pkg/
    COPY cmd/    cmd/
    COPY vendor/ vendor/
    COPY config/ config/
    RUN set -x && \
        #go get github.com/2tvenom/go-test-teamcity && \
        go get github.com/BurntSushi/toml

    # Build
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/subravi92/burrow-operator/cmd/manager

    # Copy the controller-manager into a thin image
    FROM alpine
    WORKDIR /
    COPY --from=builder /go/src/github.com/subravi92/burrow-operator/manager .
    COPY --from=builder /go/src/github.com/subravi92/burrow-operator/config config
    ENTRYPOINT ["./manager"]
    
