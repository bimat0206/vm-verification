{
  "input": "{\"verificationContext\":{\"status\":\"VERIFICATION_INITIALIZED\",\"requestMetadata\":{\"processingStarted\":\"2025-05-14T06:49:52.604Z\",\"requestTimestamp\":\"2025-05-14T06:49:52.604Z\",\"requestId\":\"2e111625-0d7a-409f-afb5-43ab0265d389\"},\"previousVerificationId\":\"\",\"verificationAt\":\"2025-05-14T06:49:52.604Z\",\"referenceImageUrl\":\"s3://kootoro-dev-s3-reference-f6d3xl/processed/2025/05/06/23591_5560c9c9_reference_image.png\",\"layoutId\":23591,\"layoutPrefix\":\"5560c9c9\",\"notificationEnabled\":false,\"checkingImageUrl\":\"s3://kootoro-dev-s3-checking-f6d3xl/AACZ 3.png\",\"turnTimestamps\":{\"initialized\":\"2025-05-14T06:49:52.604Z\"},\"verificationType\":\"LAYOUT_VS_CHECKING\",\"turnConfig\":{\"maxTurns\":2,\"referenceImageTurn\":1,\"checkingImageTurn\":2},\"vendingMachineId\":\"VM-3245\",\"verificationId\":\"4cb83621-07b7-49fc-bd15-8c53aee2d80e\"},\"schemaVersion\":\"1.0.0\"}",
  "inputDetails": {
    "truncated": false
  },
  "resource": "arn:aws:lambda:us-east-1:879654127886:function:kootoro-dev-lambda-initialize-f6d3xl"
}

{
  "output": "{\"verificationId\":\"4cb83621-07b7-49fc-bd15-8c53aee2d80e\",\"verificationAt\":\"2025-05-14T06:49:52.604Z\",\"status\":\"VERIFICATION_INITIALIZED\",\"verificationType\":\"LAYOUT_VS_CHECKING\",\"vendingMachineId\":\"VM-3245\",\"layoutId\":23591,\"layoutPrefix\":\"5560c9c9\",\"referenceImageUrl\":\"s3://kootoro-dev-s3-reference-f6d3xl/processed/2025/05/06/23591_5560c9c9_reference_image.png\",\"checkingImageUrl\":\"s3://kootoro-dev-s3-checking-f6d3xl/AACZ 3.png\",\"turnConfig\":{\"maxTurns\":2,\"referenceImageTurn\":1,\"checkingImageTurn\":2},\"turnTimestamps\":{\"initialized\":\"2025-05-14T06:49:52.604Z\"},\"requestMetadata\":{\"requestId\":\"2e111625-0d7a-409f-afb5-43ab0265d389\",\"requestTimestamp\":\"2025-05-14T06:49:52.604Z\",\"processingStarted\":\"2025-05-14T06:49:52.604Z\"},\"resourceValidation\":{\"layoutExists\":true,\"referenceImageExists\":true,\"checkingImageExists\":true,\"validationTimestamp\":\"2025-05-14T06:49:54Z\"},\"notificationEnabled\":false}",
  "outputDetails": {
    "truncated": false
  }
}