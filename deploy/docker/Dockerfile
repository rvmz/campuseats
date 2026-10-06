# Session 8: multi-stage build.
# Build stage: golang image, CGO_ENABLED=0, binary selected by ARG SERVICE (gateway, order, kitchen, notifier, migrate).
# Runtime stage: non-root user, ca-certificates, only the binary.
# Do not copy secrets or the instructor directory into the image.
