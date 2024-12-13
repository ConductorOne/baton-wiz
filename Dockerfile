FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-wiz"]
COPY baton-wiz /