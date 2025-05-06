#!/bin/bash
set -e

echo "Testing verification API..."
curl -v -X POST http://localhost:3000/api/v1/verification \
  -H "Content-Type: application/json" \
  -d '{
    "referenceImageUrl": "s3://vending-machine-verification-image-reference-a11/processed/2025/04/24/23591_54oa2052_reference_image.png",
    "checkingImageUrl": "s3://vending-machine-verification-image-checking-a12/AACZ 3.png",
    "vendingMachineId": "VM-23591",
    "layoutId": 23591,
    "layoutPrefix": "54o68a4f"
  }'

echo -e "\n\nTesting health endpoint..."
curl http://localhost:3000/health
