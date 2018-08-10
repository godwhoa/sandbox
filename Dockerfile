FROM alpine:3.7
RUN apk add --no-cache gcc libc-dev

# for testing
RUN touch /rootowned

# run `id -u && id -g` and tweak these if you run into permission issues
RUN addgroup -g 1000 -S sandbox && \
    adduser -u 1000 -S sandbox -G sandbox && \ 
    mkdir /src && \ 
    chown -R sandbox:sandbox /src

VOLUME /src/main.c
USER sandbox
CMD gcc /src/main.c -o /src/main && /src/main