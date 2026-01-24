CGO_ENABLED=1 \
GOOS=linux \
go build \
    -buildmode=pie \
    -ldflags="-linkmode=external" \
    -tags "with_utls,with_quic,netgo" \
    -o ./bin/singtoolboxgui \
    ./cmd/app/ &&
termux-elf-cleaner ./bin/singtoolboxgui