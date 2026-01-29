# DEVELOPMENT TASKS ====================================================================================================

##@ Development

.PHONY: dep
dep: ## Install dependencies
	uv venv
	uv pip install -r pyproject.toml --no-deps

.PHONY: lock
lock: ## Lock dependencies
	-rm requirements.txt
	uv pip compile pyproject.toml --generate-hashes -o requirements.txt --no-deps

.PHONY: run
run: ## Run development server
	.venv/bin/python src/manage.py runserver

.PHONY: test
test: ## Run tests
	.venv/bin/pytest --fixtures tests

##@ Code Quality

.PHONY: lint
lint: ## Run linter (ruff format + check)
	uvx ruff format
	uvx ruff check --fix .

.PHONY: typecheck
typecheck: ## Run type checker (ty)
	uvx ty check --python .venv/bin/python src/

.PHONY: check
check: lint typecheck ## Run all code quality checks

##@ Database

.PHONY: db-up
db-up: ## Start local PostgreSQL and Redis
	docker compose up -d

.PHONY: db-down
db-down: ## Stop local PostgreSQL and Redis
	docker compose down

.PHONY: db-logs
db-logs: ## Show database logs
	docker compose logs -f

.PHONY: migrate
migrate: ## Run database migrations
	.venv/bin/python src/migration.py migrate

.PHONY: makemigrations
makemigrations: ## Create new migrations
	.venv/bin/python src/migration.py makemigrations

.PHONY: dump
dump: ## Dump fixtures to JSON
	.venv/bin/python src/migration.py dumpdata goods.good > fixtures/good.json

.PHONY: restore
restore: ## Restore fixtures from JSON
	.venv/bin/python src/migration.py loaddata fixtures/good.json

##@ Static Files

.PHONY: static
static: ## Collect static files
	.venv/bin/python src/made.py collectstatic
