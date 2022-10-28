ENV_FILE :=
SHELL := /bin/bash
LICENSES_DIR := licenses

export

.PHONY: license
license:
	go mod tidy
	rm -rf "${LICENSES_DIR}"
	mkdir -p "${LICENSES_DIR}"
	go-licenses save . --force --save_path "${LICENSES_DIR}" --alsologtostderr
	chmod +w -R "${LICENSES_DIR}"

.PHONY: goreleaser
goreleaser:
	goreleaser release --snapshot --rm-dist

.PHONY: test
test:
	go test -v ./...


.PHONY: docker-rerun
docker-rerun:
	${MAKE} clean
	docker-compose build --no-cache
	docker-compose up -d
	docker-compose logs -f

.PHONY: clean
clean:
	docker-compose down
	-docker volume rm sandbox-http-server-go_postgres_data

.PHONY: run
run:
ifndef ENV_FILE
	exit 1
endif
	env $$(cat ${ENV_FILE} | xargs) go run .

.PHONY: debug
debug:
	${MAKE} run ENV_FILE=".env.debug"
