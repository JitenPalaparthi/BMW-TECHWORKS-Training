
## Basic docker commands

- docker pull nginx:alpine 

- docker run -d --name nginx-demo -p 8080:80 nginx:alpine

- docker rm -f <containerid | container-name>

- docker ps 

- docker ps -a

- docker exec -it nginx-demo sh

- open your browser localhost:8080 --> should open nginx website 

## docker build

- docker build . -t jpalaparthi/nginx-basic-demo:v0.1.0

x- docker push jpalaparthi/nginx-basic-demo:v0.1.0