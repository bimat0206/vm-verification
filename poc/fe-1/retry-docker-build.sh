#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# Silence the AWS CLI
###############################################################################
export AWS_PAGER=""                  # no interactive pager
QUIET="--no-cli-pager"               # convenience var for brevity

###############################################################################
# User‑editable configuration
###############################################################################
AWS_REGION="us-east-1"

# ECR image
ECR_REPO="879654127886.dkr.ecr.${AWS_REGION}.amazonaws.com/vending-verification-streamlit-app"

# ECS
ECS_CLUSTER="vm-fe-dev-streamlit-f6d3xl-cluster"
ECS_SERVICE="vm-fe-dev-streamlit-f6d3xl"

###############################################################################
# Derived / helper variables
###############################################################################
IMAGE_TAG=latest
IMAGE_URI="${ECR_REPO}:${IMAGE_TAG}"

echo "▶ Building & pushing ${IMAGE_URI}"

# 1. Log in, build, push (throw away noisy JSON)
aws ecr get-login-password --region "${AWS_REGION}" $QUIET \
  | docker login --username AWS --password-stdin "${ECR_REPO}"

docker build --platform linux/amd64 -t "${IMAGE_URI}" .
docker push "${IMAGE_URI}"

echo "✔ Image pushed"

# 2. Clone the current task definition
echo "▶ Cloning task definition"

CURRENT_TASK_DEF_ARN=$(aws ecs describe-services $QUIET \
  --cluster "${ECS_CLUSTER}" \
  --services "${ECS_SERVICE}" \
  --query 'services[0].taskDefinition' \
  --output text)

aws ecs describe-task-definition $QUIET \
  --task-definition "${CURRENT_TASK_DEF_ARN}" \
  --query 'taskDefinition' \
  --output json > taskdef.json

# Strip read‑only fields, swap in new image
jq '
  del(
    .taskDefinitionArn,
    .revision,
    .status,
    .requiresAttributes,
    .compatibilities,
    .registeredAt,
    .registeredBy,
    .deregisteredAt
  )
  | .containerDefinitions[0].image = "'"${IMAGE_URI}"'"
' taskdef.json > taskdef-updated.json

echo "▶ Registering new task definition"

NEW_TASK_DEF_ARN=$(aws ecs register-task-definition $QUIET \
  --cli-input-json file://taskdef-updated.json \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)

echo "✔ New task‑definition ${NEW_TASK_DEF_ARN}"

# 3. Update the ECS service (discard JSON response)
echo "▶ Updating service to new revision"

aws ecs update-service $QUIET \
  --cluster "${ECS_CLUSTER}" \
  --service "${ECS_SERVICE}" \
  --task-definition "${NEW_TASK_DEF_ARN}" \
  > /dev/null

# 4. Optionally wait for stability (this prints nothing until done)
echo "▶ Waiting for deployment to stabilise…"
aws ecs wait services-stable $QUIET \
  --cluster "${ECS_CLUSTER}" \
  --services "${ECS_SERVICE}"

echo "🎉 Deployment complete!  Service is running image ${IMAGE_TAG}"
