FROM registry.access.redhat.com/ubi8/go-toolset:1.14.12 as builder
WORKDIR /tmp/src/github.com/m-lab/alertmanager-github-receiver
ADD . ./
USER 0
RUN chown -R 1001:0 /tmp/src
USER 1001

ADD go.mod go.sum ./
RUN go mod download

# TODO(soltesz): Use vgo for dependencies.
ENV CGO_ENABLED 0


RUN go build \
       -v \
      -ldflags "-X github.com/m-lab/go/prometheusx.GitShortCommit=$(git log -1 --format=%h)" \
       ./cmd/github_receiver

FROM registry.access.redhat.com/ubi8/ubi-micro
WORKDIR /app
COPY --from=builder /tmp/src/github.com/m-lab/alertmanager-github-receiver/github_receiver ./

EXPOSE 9393

ENTRYPOINT ["/app/github_receiver"]
