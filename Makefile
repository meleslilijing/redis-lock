test:
	@go test -v

tidy:
	@go mod tidy -v

ut:
	@go test -v

e2e:
	sh ./script/integrate_test.sh
    
e2e_up:
	docker compose -f script/docker-compose.yml up -d

e2e_down:
	docker compose -f script/docker-compose.yml down

mock:
	mockgen -copyright_file=.license.go.header -package=mocks -destination=mocks/redis_cmdable.mock.go github.com/redis/go-redis/v9 Cmdable