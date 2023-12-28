FROM golang:1.21.3 as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /go/bin/risken-review main.go

# FROM python:3.9-alpine
# RUN apk add --no-cache gcc musl-dev libffi-dev make git
# RUN pip install semgrep==1.46.0
FROM returntocorp/semgrep:1.46.0
COPY --from=builder /go/bin/risken-review /usr/local/bin/
RUN apk add git
WORKDIR /usr/local/bin
ENV \
  GITHUB_TOKEN= \
  GITHUB_EVENT_PATH= \
  GITHUB_WORKSPACE=
ENTRYPOINT [ "risken-review" ]
