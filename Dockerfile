from golang:1.11.2 as build
add . /src
run cd /src && CGO_ENABLED=0 go get ./cmd/webps-entrypoint

from gcr.io/distroless/base

# Copy over binaries from the build.
copy --from=build /go/bin/webps-entrypoint /bin/

entrypoint ["/bin/webps-entrypoint"]
