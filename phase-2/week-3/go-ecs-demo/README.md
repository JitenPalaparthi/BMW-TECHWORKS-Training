aws ecr get-login-password --region eu-north-1 \
| docker login \
  --username AWS \
  --password-stdin \
  $(aws sts get-caller-identity --query Account --output text).dkr.ecr.eu-north-1.amazonaws.com