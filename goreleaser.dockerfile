FROM scratch
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/gosh /usr/bin/gosh
ENTRYPOINT ["/usr/bin/gosh"]
