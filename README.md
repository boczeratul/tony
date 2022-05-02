# portto

Portto backend interview assignment

## Installation

Build the API server docker image

```sh
# At the root directory of this repo
make build
```

## How to Run

### API Server

After installation, run up API server and MySQL server \
***It might take a while to start up***

```sh
# At the root directory of this repo
make run
```

### Indexer

Index the most recent block only

```sh
docker run --network=host portto-indexer:1.0-alpine
```

or specify a block number, 18952359 for example

```sh
docker run --network=portto_portto --entrypoint=/bin/sh portto-indexer:1.0-alpine -c "/go/bin/main --sqlHost=mysql --blockNumber=18952359"
```

## Test

### Get blocks

Default limit = 20

- Without limit query parameter

```sh
curl --location --request GET 'localhost:3000/blocks' \
--data-raw ''
```

- With limit query parameter

```sh
curl --location --request GET 'localhost:3000/blocks?limit=5' \
--data-raw ''
```

### Get block by Id

```sh
curl --location --request GET 'localhost:3000/blocks/0x8848670eef090a03bef2ccc3ad634eb001541f2dd9832d2b387140af05658894'
```

### Get transaction by hash

```sh
curl --location --request GET 'localhost:3000/transaction/0xd515fdbefad7e12cbb16f3f554a23e6f741c08924992108b10530efbdf9589bc'
```
