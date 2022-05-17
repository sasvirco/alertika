FROM golang:1.16 as build-env

WORKDIR /
ADD . /
RUN make

FROM gcr.io/distroless/base

# changed alertika.linux.amd64 to alertika.linux.arm64 for graviton processors
COPY --from=build-env /bin/alertika.linux.amd64 /
CMD ["/alertika.linux.amd64"] 
