ENV_FILE :=
SHELL := /bin/bash
PRODUCTION_BIN_VERSION := 0.0.0

export

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
	docker-compose build
	docker push pollenjp/pomodoro-bot:latest
