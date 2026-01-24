gogio \
    -target android \
    -arch arm64 \
    -o ./bin/SingToolBoxGUI.apk \
    -icon ./android/play_store_512.png \
    -appid com.bghorizon.singtoolboxgui \
    -name SingToolBoxGUI \
    -tags "with_utls,with_quic" \
    ./cmd/app/