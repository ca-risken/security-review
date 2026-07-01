FROM golang:1.25.5 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /go/bin/risken-review main.go

FROM semgrep/semgrep:1.168.0
COPY --from=builder /go/bin/risken-review /usr/local/bin/
WORKDIR /usr/local/bin
ENV \
  GITHUB_TOKEN= \
  GITHUB_EVENT_PATH= \
  GITHUB_WORKSPACE=
ENTRYPOINT [ "risken-review" ]
