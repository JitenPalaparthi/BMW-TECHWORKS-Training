# Copy files from host to the container

```bash
docker cp . nginx-demo:/usr/share/nginx/html/
```
# build docker image

```bash
docker build -f mydockerfile . -t jpalaparthi/html-forms-demo:v0.1.0
```

# run container

```bash
docker run -d -p 8090:80 --name nginx-form-demo jpalaparthi/html-forms-demo:v0.1.0 
```