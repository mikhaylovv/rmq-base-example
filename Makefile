.PHONY: basic confirms cleanup

basic:
	docker compose --profile basic up --build

confirms:
	docker compose --profile confirms up --build

cleanup:
	docker compose --profile basic --profile confirms down
