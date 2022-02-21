#!/bin/bash
set -e
CLOUDFRONT_DISTRIBUTION="E1DAF3Z7T3PIYW"

aws s3 sync --debug --region us-west-1 ./static/ s3://code.stevearc.com/
aws cloudfront create-invalidation --distribution-id "$CLOUDFRONT_DISTRIBUTION" --paths '/*'
