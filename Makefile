image:
	docker build . -t rewrite

run-echo:
	go run . -c examples/echo.yaml

run:
	docker run -it --rm --mount type=bind,source=`pwd`/examples/echo.yaml,target=/config.yaml -p 0.0.0.0:8081:8081 rewrite

run-2:
	docker run -it --rm --mount type=bind,source=`pwd`/examples/echo.yaml,target=/config.yaml -p 0.0.0.0:8081:8081 lindex/http-rewrite:v0.3.0