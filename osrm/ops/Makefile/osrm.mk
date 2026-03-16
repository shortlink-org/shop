# OSRM TASKS ===========================================================================================================

# Variables
DATA_DIR := $(SELF_DIR)data
OSM_FILE := berlin-latest.osm.pbf
OSRM_IMAGE := osrm/osrm-backend
OSRM_CONTAINER := osrm-berlin
OSRM_PORT := 5000

##@ OSRM Graph Generation

.PHONY: osrm-extract
osrm-extract: ## Step 1: Extract OSM data (creates .osrm files)
	docker run -t --rm -v $(DATA_DIR):/data $(OSRM_IMAGE) \
		osrm-extract -p /opt/car.lua /data/$(OSM_FILE)

.PHONY: osrm-partition
osrm-partition: ## Step 2: Partition the graph (MLD algorithm)
	docker run -t --rm -v $(DATA_DIR):/data $(OSRM_IMAGE) \
		osrm-partition /data/berlin-latest.osrm

.PHONY: osrm-customize
osrm-customize: ## Step 3: Customize the graph
	docker run -t --rm -v $(DATA_DIR):/data $(OSRM_IMAGE) \
		osrm-customize /data/berlin-latest.osrm

.PHONY: osrm-build
osrm-build: osrm-extract osrm-partition osrm-customize ## Build OSRM graph (all steps)
	@echo "✅ OSRM graph built successfully!"
	@echo "   Run 'make osrm-run' to start the routing server"

##@ OSRM Server

.PHONY: osrm-run
osrm-run: ## Start OSRM routing server on port 5000
	docker run -d --name $(OSRM_CONTAINER) -p $(OSRM_PORT):5000 \
		-v $(DATA_DIR):/data $(OSRM_IMAGE) \
		osrm-routed --algorithm=MLD /data/berlin-latest.osrm
	@echo "✅ OSRM server started at http://localhost:$(OSRM_PORT)"
	@echo "   Test: curl 'http://localhost:$(OSRM_PORT)/route/v1/driving/13.388860,52.517037;13.397634,52.529407?overview=full'"

.PHONY: osrm-stop
osrm-stop: ## Stop OSRM routing server
	docker stop $(OSRM_CONTAINER) && docker rm $(OSRM_CONTAINER)
	@echo "✅ OSRM server stopped"

.PHONY: osrm-status
osrm-status: ## Check OSRM server status
	@docker ps --filter name=$(OSRM_CONTAINER) --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "OSRM server not running"

.PHONY: osrm-test
osrm-test: ## Test OSRM routing API
	@curl -s 'http://localhost:$(OSRM_PORT)/route/v1/driving/13.388860,52.517037;13.397634,52.529407?overview=full' | jq '.routes[0] | {distance, duration}'

##@ OSRM Cleanup

.PHONY: osrm-clean
osrm-clean: ## Remove generated OSRM files (keeps .osm.pbf)
	rm -f $(DATA_DIR)/*.osrm*
	@echo "✅ OSRM files cleaned"
