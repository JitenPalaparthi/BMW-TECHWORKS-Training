docker compose up -d 

docker compose down -v

docker build . -t jpalaparthi/demo-java-crud:v0.0.1

docker run -d --name demo-app -p 8080:8080 jpalaparthi/demo-java-crud:v0.0.1