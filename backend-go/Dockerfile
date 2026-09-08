# Multi-stage build: the toolchain lives in the first stage, the final image
# gets nothing but the compiled binary.

# ---------- stage 1: build ----------
FROM golang:1.25 AS build

# Never download a toolchain during the build. If go.mod ever requires a
# newer Go than this base image has, fail loudly instead of silently
# reaching out to the network mid-build.
ENV GOTOOLCHAIN=local

WORKDIR /src

# Dependencies first, in their own layer: it is rebuilt only when go.mod or
# go.sum change, not on every source edit.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 links everything statically, so the binary needs no libc and
# can run in an image that contains literally nothing else.
# -ldflags="-s -w" strips the symbol table and DWARF debug info (~25% smaller).
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/backend-go .

# ---------- stage 2: runtime ----------
# `scratch` is the empty image: no shell, no package manager, no libc.
# Nothing to exploit and nothing to patch, at the cost of not being able to
# `docker exec` into it for debugging.
FROM scratch

# Numeric UID, not a name: there is no /etc/passwd in scratch to resolve one.
# 65534 is the conventional "nobody".
USER 65534:65534

COPY --from=build /out/backend-go /backend-go

EXPOSE 8000

ENTRYPOINT ["/backend-go"]
