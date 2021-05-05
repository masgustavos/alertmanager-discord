# Built following https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324

# STEP 1 build executable binary
FROM golang:alpine as builder
# Install SSL ca certificates
RUN apk update && apk add git && apk add ca-certificates
# Create appuser
RUN adduser -D -g '' appuser
COPY . $GOPATH/src/masgustavos/alertmanager-discord/
WORKDIR $GOPATH/src/masgustavos/alertmanager-discord/
#get dependancies
RUN go get -d -v
#build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/alertmanager-discord

# STEP 2 build a small image
# start from scratch
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable
COPY --from=builder /go/bin/alertmanager-discord /go/bin/alertmanager-discord

ENV LISTEN_ADDRESS=0.0.0.0:8080
EXPOSE 8080
USER appuser
ENTRYPOINT ["/go/bin/alertmanager-discord"]