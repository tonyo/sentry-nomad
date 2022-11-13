FROM golang:1.19 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/sentry-nomad

# Now copy it into the distrolessbase image: https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/sentry-nomad /
CMD ["/sentry-nomad"]
