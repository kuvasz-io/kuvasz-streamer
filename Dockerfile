FROM scratch

LABEL maintainer="kuvasz.io  <info@kuvasz.io>"

COPY ./kuvasz-streamer  /

ENTRYPOINT ["/kuvasz-streamer"]
