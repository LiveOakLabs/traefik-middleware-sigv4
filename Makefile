.PHONY: lint test vendor clean

export GO111MODULE=on

default: lint test

lint:
	golangci-lint run

test:
	go test -v -cover ./...

yaegi_test:
	yaegi test -v .

vendor:
	go mod vendor

clean:
	rm -rf ./vendor

.PHONY: mc
mc:
	mc mb -p play/treafikmiddlewares3v4sig
	echo "<h1>hi</h1>" | mc pipe play/treafikmiddlewares3v4sig/index.html
