# Utility Knife

This is a simple Go web server that can perform a number of simple, common tasks that may be useful to your learning and development using various linux container technologies (such as Docker, Kubernetes, etc.)

## Features

* Simple endpoints for returning data such as the server hostname and IP address fulfilling the request. There are HTML and JSON endpoints to return said data.
* An echo endpoint to inspect the contents of your request.
* A data endpoint which acts as a simple key/value store. Data is stored in `/var/lib/utilityknife`. This is useful if you need to play around with various volume/storage/persistent storage features.

## Usage

If you intend to use this with Docker or a Docker-related technology like Kubernetes, there already exists a Docker image for you to use: [weters/utilityknife](https://hub.docker.com/r/weters/utilityknife).

```
$ docker container run --rm --name utilityknife -p 8080:80 weters/utilityknife:latest
```

If you want to build it yourself, it can be built using Go:
```
$ go get -u github.com/weters/utilityknife
$ $GOPATH/bin/utilityknife -addr :8080
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

### Echo Endpoint

The echo endpoint is useful if you want to inspect how your requests look from the context of a server.

Example:

```
$ curl -v '127.0.0.1/echo/foo/bar' -H 'Content-Type: application/json' -d '{"status":"OK"}'
> POST /echo/foo/bar HTTP/1.1
> Host: 127.0.0.1
> Content-Type: application/json
>
< HTTP/1.1 200 OK
< Content-Type: text/plain; charset=utf-8
< X-Hostname: utilityknife-6499f48f79-tnpzd
< X-Ip: 10.1.0.73
< X-Served-By: weters/utilityknife
<
POST /echo/foo/bar HTTP/1.1
Host: 127.0.0.1
Accept: */*
Content-Length: 15
Content-Type: application/json
User-Agent: curl/7.54.0

{"status":"OK"}
```

### Data Endpoint

The data endpoint can be used as a simple key/value store. All data is stored in `/var/lib/utilityknife`. Locking is only guaranteed for a single replica, so if you have multiple replicas access the same volume, you may (but probably won't) run into race conditions if you are doing heavy read/writes to the same keys.

Example:

```
$ curl -v '127.0.0.1/data/key1' -d 'bar' -X PUT
> PUT /data/key1 HTTP/1.1
> Host: 127.0.0.1
> Content-Type: application/x-www-form-urlencoded
>
< HTTP/1.1 201 Created
< X-Hostname: utilityknife-6499f48f79-tl4b5
< X-Ip: 10.1.0.74
< X-Served-By: weters/utilityknife
 
$ curl -v '127.0.0.1/data/key2' -d '["bar1","bar2"]' -H 'Content-Type: application/json' -X PUT
> PUT /data/key2 HTTP/1.1
> Host: 127.0.0.1
> Content-Type: application/json
>
< HTTP/1.1 201 Created
< X-Hostname: utilityknife-6499f48f79-tl4b5
< X-Ip: 10.1.0.74
< X-Served-By: weters/utilityknife

$ curl -v '127.0.0.1/data/key1'
> GET /data/key1 HTTP/1.1
> Host: 127.0.0.1
>
< HTTP/1.1 200 OK
< Content-Type: application/x-www-form-urlencoded
< X-Hostname: utilityknife-6499f48f79-tnpzd
< X-Ip: 10.1.0.73
< X-Served-By: weters/utilityknife
bar%

$ curl -v '127.0.0.1/data/key2'
> GET /data/key2 HTTP/1.1
> Host: 127.0.0.1
>
< HTTP/1.1 200 OK
< Content-Type: application/json
< X-Hostname: utilityknife-6499f48f79-tnpzd
< X-Ip: 10.1.0.73
< X-Served-By: weters/utilityknife
["bar1","bar2"]%
````
