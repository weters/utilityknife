# Utility Knife

This is a simple Go web server that can perform a number of simple, common tasks that may be useful to your learning and development using various linux container technologies (such as Docker, Kubernetes, etc.)

## Usage

If you intend to use this with Docker or a Docker-related technology like Kubernetes, there already exists a Docker image for you to use: [weters/utilityknife](https://hub.docker.com/r/weters/utilityknife).

```
> docker container run --rm --name utilityknife -p 8080:80 weters/utilityknife:latest
```

If you want to build it yourself, it can be built using Go:
```
> go get -u github.com/weters/utilityknife
> $GOPATH/bin/utilityknife -addr :8080
```

## Endpoints

Method | Endpoint | Description
--- | --- | ---
`GET` | `/` | Returns a simple HTML page with server details such as hostname, IP and server date/time
`GET` | `/json` | Similar to `/` but returns the data in a JSON format
`*` | `/echo/*` | Returns the request as a `text/plain` object in the response
`PUT` | `/data/*` | Will store any data submitted to the server under the key defined by the URL path
`GET` | `/data/*` | Will return any data stored under the key defined by the URL path
`DELETE` | `/data/*` | Will delete any data stored under the key defined by the URL path
