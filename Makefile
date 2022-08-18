ENV_FILE :=

export

.PHONY: run
run:
ifndef ENV_FILE
	exit 1
endif
	env $$(cat ${ENV_FILE} | xargs) go run .

.PHONY: debug
debug:
	${MAKE} run ENV_FILE=".env"

.PHONY: production
production:
	${MAKE} run ENV_FILE=".env.production"
