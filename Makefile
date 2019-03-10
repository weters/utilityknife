VERSION ?= "1.0"

docker:
	docker build -t weters/utilityknife .

docker-tag: docker
	docker tag weters/utilityknife weters/utilityknife:${VERSION}

docker-push:
	docker push weters/utilityknife

docker-tag-push: docker-tag docker-push

.PHONY: docker docker-tag docker-tag-push
