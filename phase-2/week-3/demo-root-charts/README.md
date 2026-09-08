

helm package demo-root-charts

helm repo index . \
  --url https://jitenpalaparthi.github.io/demo-root-charts

helm repo add jiten-demo \
  https://jitenpalaparthi.github.io/demo-root-charts

helm install demo-release \
  jiten-demo/demo-root-charts


.nojekyll
index.yaml
demo-root-charts-0.1.0.tgz
demo-root-charts-0.1.2.tgz

helm repo add helm-demo https://aws-helm-charts-s3.s3.ap-south-1.amazonaws.com/

aws s3

https://aws-helm-charts-s3.s3.ap-south-1.amazonaws.com/

{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadHelmRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::aws-helm-charts-s3/*"
        }
    ]
}

