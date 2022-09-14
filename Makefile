ENV_FILE :=
SHELL := /bin/bash
PRODUCTION_BIN_VERSION := 0.1.0
LICENSES_DIR := "licenses"

export

.PHONY: license
license:
	rm -rf "${LICENSES_DIR}"
	mkdir -p "${LICENSES_DIR}"
	go-licenses save . --force --save_path "${LICENSES_DIR}" --alsologtostderr
	chmod +w -R "${LICENSES_DIR}"

.PHONY: goreleaser
goreleaser:
	goreleaser release --snapshot --rm-dist

.PHONY: run
run:
ifndef ENV_FILE
	exit 1
endif
	env $$(cat ${ENV_FILE} | xargs) go run .

.PHONY: debug
debug:
	${MAKE} run ENV_FILE=".env.debug"

.PHONY: production
production:
	${MAKE} run ENV_FILE=".env.production"

.PHONY: run-bin
run-bin:
ifndef ENV_FILE
	exit 1
endif
	env $$(cat ${ENV_FILE} | xargs) ./pomodoro-bot

.PHONY: production-bin
production-bin:
	if ![[ -f "pomodoro-bot.tar.gz" ]]; then\
		wget -O "pomodoro-bot.tar.gz" \
			"https://github.com/pollenjp/pomodoro-bot/releases/download/v${PRODUCTION_BIN_VERSION}/pomodoro-bot_${PRODUCTION_BIN_VERSION}_linux_amd64.tar.gz" ;\
	fi
	tar -xvf pomodoro-bot.tar.gz
	${MAKE} run-bin ENV_FILE=".env.production"

.PHONY: docker
docker:
	docker-compose build --no-cache
	docker push pollenjp/pomodoro-bot:latest
	docker image inspect pollenjp/pomodoro-bot:latest \
		--format '{{ (index (split .Id ":") 1) }}' \
		| xargs -I{} \
		docker tag \
			{} pollenjp/pomodoro-bot:v${PRODUCTION_BIN_VERSION}
	docker push pollenjp/pomodoro-bot:v${PRODUCTION_BIN_VERSION}
