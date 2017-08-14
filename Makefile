build-feed:
	go build -o bin/feed ./feed

run-feed:
	make build-feed && ./bin/feed

build-importer:
	go build -o bin/event_importer ./event_importer

run-importer:
	make build-importer && ./bin/event_importer

build-mock-store:
	go build -o bin/mock_store ./mock_store

run-mock-store:
	make build-mock-store && ./bin/mock_store
