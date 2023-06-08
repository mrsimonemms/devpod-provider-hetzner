FROM golang AS builder
ARG GIT_COMMIT
ARG GIT_REPO
ARG VERSION
WORKDIR /app
ADD . .
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV PROJECT_NAME="${PROJECT_NAME}"
RUN go build \
  -ldflags \
  "-w -s -X $GIT_REPO/cmd.Version=$VERSION -X $GIT_REPO/cmd.GitCommit=$GIT_COMMIT" \
  -o /app/app
ENTRYPOINT /app/app

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/app /app
ENTRYPOINT [ "/app/app" ]
