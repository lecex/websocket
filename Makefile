.PHONY: git
git:
	git add .
	git commit -m"自动提交 git 代码"
	git push
.PHONY: tag
tag:
	git push --tags
.PHONY: rpc
rpc:
	micro api  --handler=rpc  --namespace=go.micro.api --address=:8080
.PHONY: api
api:
	micro api  --handler=api  --namespace=go.micro.api --address=:8081
.PHONY: web
web:
	micro api  --handler=web  --namespace=go.micro.web --address=:8082

.PHONY: proto
proto:

.PHONY: docker
docker:
	docker build -f Dockerfile  -t websocket .
.PHONY: run
run:
	go run main.go
test:
	go test main_test.go -test.v
t:
	git tag -d v1.3.3
	git push origin :refs/tags/v1.3.3
	git tag v1.3.3
	git push --tags