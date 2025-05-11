#!/usr/bin/env bash

# Set environment variables for testing
export REFERENCE_BUCKET=kootoro-dev-s3-reference-x1y2z3
export CHECKING_BUCKET=kootoro-dev-s3-checking-f6d3xl

# Build the application
echo "Building application..."
go build -o validation-test cmd/main.go

# Test with correct URLs
echo -e "\n\nTesting with correct URLs..."
cat <<EOF > /tmp/correct-test.json
{
  "verificationContext": {
    "verificationId": "verif-20240511-123456",
    "verificationAt": "2024-05-11T12:34:56Z",
    "status": "INITIATED",
    "verificationType": "LAYOUT_VS_CHECKING",
    "vendingMachineId": "VM-12345",
    "layoutId": 12345,
    "layoutPrefix": "layout_",
    "referenceImageUrl": "s3://kootoro-dev-s3-reference-x1y2z3/path/to/reference.jpg",
    "checkingImageUrl": "s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg"
  },
  "layoutMetadata": {
    "machineStructure": {
      "rowCount": 5,
      "columnsPerRow": 8,
      "rowOrder": ["A", "B", "C", "D", "E"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8"]
    }
  }
}
EOF

./validation-test < /tmp/correct-test.json

# Test with incorrect reference bucket
echo -e "\n\nTesting with incorrect reference bucket..."
cat <<EOF > /tmp/incorrect-reference.json
{
  "verificationContext": {
    "verificationId": "verif-20240511-123456",
    "verificationAt": "2024-05-11T12:34:56Z",
    "status": "INITIATED",
    "verificationType": "LAYOUT_VS_CHECKING",
    "vendingMachineId": "VM-12345",
    "layoutId": 12345,
    "layoutPrefix": "layout_",
    "referenceImageUrl": "s3://wrong-bucket/path/to/reference.jpg",
    "checkingImageUrl": "s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg"
  },
  "layoutMetadata": {
    "machineStructure": {
      "rowCount": 5,
      "columnsPerRow": 8,
      "rowOrder": ["A", "B", "C", "D", "E"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8"]
    }
  }
}
EOF

./validation-test < /tmp/incorrect-reference.json

# Test with incorrect checking bucket (the issue we're trying to fix)
echo -e "\n\nTesting with incorrect checking bucket..."
cat <<EOF > /tmp/incorrect-checking.json
{
  "verificationContext": {
    "verificationId": "verif-20240511-123456",
    "verificationAt": "2024-05-11T12:34:56Z",
    "status": "INITIATED",
    "verificationType": "LAYOUT_VS_CHECKING",
    "vendingMachineId": "VM-12345",
    "layoutId": 12345,
    "layoutPrefix": "layout_",
    "referenceImageUrl": "s3://kootoro-dev-s3-reference-x1y2z3/path/to/reference.jpg",
    "checkingImageUrl": "s3://wrong-bucket/path/to/checking.jpg"
  },
  "layoutMetadata": {
    "machineStructure": {
      "rowCount": 5,
      "columnsPerRow": 8,
      "rowOrder": ["A", "B", "C", "D", "E"],
      "columnOrder": ["1", "2", "3", "4", "5", "6", "7", "8"]
    }
  }
}
EOF

./validation-test < /tmp/incorrect-checking.json

# Clean up
rm /tmp/correct-test.json /tmp/incorrect-reference.json /tmp/incorrect-checking.json
rm validation-test