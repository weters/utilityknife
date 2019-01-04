FROM golang:latest AS build
COPY . /go/src/github.com/weters/utilityknife
RUN go get github.com/weters/utilityknife \
 && CGO_ENABLED=0 go install github.com/weters/utilityknife

FROM busybox:latest
EXPOSE 80
COPY --from=build /go/bin/utilityknife /bin/utilityknife
VOLUME /var/lib/utilityknife
ENTRYPOINT [ "/bin/utilityknife" ]
