@echo off
go build ^
    -tags "with_utls,with_quic" ^
    -o ./bin/singtoolboxgui.exe ^
    ./cmd/app/