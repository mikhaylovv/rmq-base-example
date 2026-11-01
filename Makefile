.PHONY: basic confirms publisher clean

basic:
	docker compose --profile basic up --build

confirms:
	docker compose --profile confirms up --build

publisher:
	docker compose --profile publisher up --build

clean:
	docker compose --profile basic --profile confirms --profile publisher down
