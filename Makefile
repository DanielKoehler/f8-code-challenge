build-feed:
	go build -o bin/feed github.com/fresh8/f8-code-challenge/feed

run-feed:
	make build-feed && ./bin/feed
