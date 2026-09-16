curl -X POST \
https://demo-go.eu-north-1.elasticbeanstalk.com//api/users \
-H "Content-Type: application/json" \
-d '{
  "name": "Jiten",
  "email": "jiten@example.com"
}'

http://demo-go.eu-north-1.elasticbeanstalk.com/