Version := beta
default:
	@echo 'Usage of make: [ build | linux_build | windows_build | clean ]'
coveralls:
	@#go install github.com/mattn/goveralls@latest
	@#go install golang.org/x/tools/cmd/cover@latest
	#https://coveralls.io/
	@go test ./... -v -covermode=count -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@#goveralls -coverprofile=coverage.out -service=travis-ci -repotoken 3azIdp3c09dMBBbsrbs3l6qK1mGuEUowo
codebeat:
	@#https://codebeat.co/
goreportcard:
	@#https://goreportcard.com/report/github.com/fghosth/utils
init:
	@git init
	@chmod +x .githooks/go_pre_commit.sh
	@cp .githooks/pre-commit .git/hooks && chmod +x ./.git/hooks/pre-commit && go mod tidy && go mod vendor
	@cp .githooks/commit-msg .git/hooks && chmod +x ./.git/hooks/commit-msg
gitlog:
	@#brew tap git-chglog/git-chglog    #https://github.com/git-chglog/git-chglog
	@#brew install git-chglog
	@git-chglog -o ./docs/CHANGELOG.md
install: build
	@mv ./dist/server $(GOPATH)/dist/

clean:
	@rm -rf ./dist
htmlStatic:
	@#statik -src=/Users/derek/project/devops/valhalla/docs
	@statik -p=statik -dest=/Users/derek/project/devops/valhalla -src=/Users/derek/project/devops/valhalla/source
Debug:
	@go tool pprof -inuse_space http://127.0.0.1:6060/debug/pprof/heap #输入 inuse_space切换到常驻内存，alloc_space切换到分配的临时内存
	@go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
goconvey:
	@goconvey
mockgen:
	@mockgen -destination=mocks/go -package=mocks -source=filex/file.go