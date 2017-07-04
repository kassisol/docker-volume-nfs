FROM alpine

RUN apk --update --no-cache add \
		ca-certificates \
		libnfs \
		nfs-utils \
	&& mkdir -p /var/lib/docker/volumes /var/lib/docker/state

COPY build/docker-volume-nfs /docker-volume-nfs

ENTRYPOINT ["/docker-volume-nfs"]
