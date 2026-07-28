docker compose up -d 

- To remove all compose containers
docker compose down -v

- Build an image
docker build . -t jpalaparthi/demo-java-crud:v0.0.1

- To push an image to container registry 

docker push jpalaparthi/demo-java-crud:v0.0.1

- Run a container
docker run -d --name demo-app -p 8080:8080 jpalaparthi/demo-java-crud:v0.0.1

- remove a container (forcely)
docker rm -f demo-app

