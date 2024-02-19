FROM scratch

LABEL maintainer="kuvasz.io  <info@kuvasz.io>"

COPY ./kuvasz-streamer  /

CMD /kuvasz-streamer
