# Build the manager binary
FROM docker.io/library/golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
#COPY main.go main.go
#COPY api/ api/
#COPY controllers/ controllers/
COPY . .
COPY .git/ ./.git/
COPY .gitignore ./.gitignore

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags . -o manager
RUN go version -m ./manager > manager-version


FROM scratch
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/manager-version .
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/manager"]
