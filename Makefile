build:
	goreleaser build --snapshot --clean --verbose
clean:
	rm -rf ./dist
