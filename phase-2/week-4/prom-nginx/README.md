docker run -d \
  --name nginx \
  -p 8080:80 \
  -v "$(pwd)/nginx.conf:/etc/nginx/nginx.conf:ro" \
  nginx

docker network create monitoring

docker run -d \
  --name nginx-exporter \
  --network monitoring \
  -p 9113:9113 \
  nginx/nginx-prometheus-exporter \
  --nginx.scrape-uri=http://nginx/stub_status

docker run -d \
  --name prometheus \
  --network monitoring \
  -p 9090:9090 \
  -v "$(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
  prom/prometheus

  # simulate the traffic 

  for i in {1..100}; do

    curl -s http://localhost:8080 > /dev/null

    done