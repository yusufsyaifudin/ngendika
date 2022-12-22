CURRENT_TIME := `date +%s`

generate-swagger:
	go run main.go apidoc
	swagger validate assets/swaggerui/swagger.json
#	rm -rf internal/httpclient
#	mkdir -p internal/httpclient
#	swagger generate client -f ./docs/api.swagger.json -t internal/httpclient -A DanBam

generate-proto:
	protoc -I=./proto --go_out=plugins=grpc:./proto proto/*.proto

generate-cli-doc:
	go run main.go docs

# make create-migration NAME="create_users_table"
create-migration:
	@[ ! -z ${NAME} ]
	mkdir -p assets/migrate
	python3 scripts/makefile_helper/helper.py write_migration ${NAME}
	@go fmt assets/migrate/*

test:
	@echo "=================================================================================="
	@echo "Coverage Test"
	@echo "=================================================================================="
	go fmt ./... && go test -race -coverprofile coverage.cov -cover ./... # use -v for verbose
	@echo "\n"
	@echo "=================================================================================="
	@echo "All Package Coverage"
	@echo "=================================================================================="
	go tool cover -func coverage.cov