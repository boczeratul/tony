run:
	docker-compose up &

build:
	 DOCKER_BUILDKIT=1 docker build --progress=plain -f ./Dockerfile -t portto-api:1.0-alpine ./src
	 DOCKER_BUILDKIT=1 docker build --progress=plain -f ./indexer.Dockerfile -t portto-indexer:1.0-alpine ./src

clean:
	docker-compose down