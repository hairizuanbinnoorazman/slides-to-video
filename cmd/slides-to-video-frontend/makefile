rebuild:
	elm make src/* --output main.js
	docker stop frontend && docker rm frontend || true
	docker build -t frontend .
	docker run -p 8081:8080 --name=frontend -e SERVER_ENDPOINT=http://localhost:8880 frontend