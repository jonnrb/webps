from golang:1.10.1 as build
workdir /go/src/go.jonnrb.io/webps

# Copy over and build from the bottom of the dependency DAG up.
copy pb  ./pb
run go get -v -d ./pb/...
copy psk ./psk
run go get -v -d ./psk/...
# be and fe depend on psk and pb
copy be ./be
run go get -v -d ./be/...
copy fe  ./fe
run go get -v -d ./fe/...
# cmd/... depends on everything.
copy cmd ./cmd
run go get -v -d ./cmd/... \
 && go build -v -ldflags "-w" ./cmd/... \
 && go install -v -ldflags "-w" ./cmd/webps-entrypoint

from gcr.io/distroless/base

# Copy over binaries from the build.
copy --from=build /go/bin/* /bin/

# Copy template and static.
workdir /srv
copy srv .

entrypoint ["/bin/webps-entrypoint"]
