FROM gcr.io/distroless/static:nonroot

# `nonroot` coming from distroless
USER 65532:65532

COPY iam-runtime-infratographer /bin/iam-runtime-infratographer

# Run the runtime service on container startup.
ENTRYPOINT ["/bin/iam-runtime-infratographer"]

CMD ["serve"]
