@PHONY: run
run:
	@echo Running server...
	@go run cmd/api/main.go -port=8080 -environment=production

@PHONY: curl/health
curl/health:
	@curl http://localhost:8080/v1/healthcheck
