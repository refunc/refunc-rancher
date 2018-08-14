FROM alpine:3.7

LABEL maintainer="antmanler(wo@zhaob.in)"

RUN apk --no-cache add ca-certificates bash wget curl

COPY refunc-rancher /usr/bin/
CMD ["refunc-rancher"]
