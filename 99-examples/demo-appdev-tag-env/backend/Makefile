__DOCKER=$(shell which docker)

lint:
	goreturns -w main.go cmd internal configs

clean:
	$(__DOCKER) system prune -f
