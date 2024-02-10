FROM jammy

LABEL maintainer="kuvasz.io  <info@kuvasz.io>"

RUN mkdir -p /log
    
COPY conf       /conf

COPY ./kuvasz-streamer  /

CMD NODENAME=`cat /etc/host_hostname` /kuvasz-streamer
