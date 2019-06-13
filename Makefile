up:
	docker-compose up -d --build

frontend/build:
	docker build -t frontend:dev frontend
