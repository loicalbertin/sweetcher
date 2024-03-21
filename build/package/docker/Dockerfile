FROM golang:1 as build-env

WORKDIR /go/src/app
COPY  . ./
RUN CGO_ENABLED=0 go build -o /go/src/app/bin/sweetcher

FROM gcr.io/distroless/static

COPY --from=build-env /go/src/app/bin/sweetcher /sweetcher
COPY ./build/package/docker/sweetcher.yaml /etc/sweetcher/sweetcher.yaml

ENTRYPOINT [ "/sweetcher"]
CMD ["serve"]
