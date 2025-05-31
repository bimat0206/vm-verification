# modules/api_gateway/models.tf

# Common Models
resource "aws_api_gateway_model" "error" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "${replace(var.stage_name, "-", "_")}error"
  description  = "Error response model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      error = {
        type = "string"
      }
      message = {
        type = "string"
      }
      details = {
        type = "object"
      }
    }
    required = ["error", "message"]
  })
}

resource "aws_api_gateway_model" "empty" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "${replace(var.stage_name, "-", "_")}empty"
  description  = "Empty model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
  })
}

# Verification Models
resource "aws_api_gateway_model" "verification_request" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "VerificationRequest"
  description  = "Verification request model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      verificationContext = {
        type = "object"
        properties = {
          verificationType = {
            type = "string",
            enum = ["LAYOUT_VS_CHECKING", "PREVIOUS_VS_CURRENT"]
          }
          referenceImageUrl = {
            type = "string"
          }
          checkingImageUrl = {
            type = "string"
          }
          notificationEnabled = {
            type = "boolean"
          }
        }
        required = ["verificationType", "referenceImageUrl", "checkingImageUrl", "notificationEnabled"]
      }
    }
    required = ["verificationContext"]
  })
}

resource "aws_api_gateway_model" "verification_result" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "VerificationResult"
  description  = "Verification result model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      verificationId = {
        type = "string"
      }
      verificationAt = {
        type = "string"
      }
      status = {
        type = "string"
      }
      result = {
        type = "object"
      }
      resultImageUrl = {
        type = "string"
      }
    }
    required = ["verificationId", "verificationAt", "status"]
  })
}

resource "aws_api_gateway_model" "verification_list" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "VerificationList"
  description  = "Verification list model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      results = {
        type = "array"
        items = {
          type = "object"
          properties = {
            verificationId = {
              type = "string"
            }
            verificationAt = {
              type = "string"
            }
            vendingMachineId = {
              type = "string"
            }
            verificationStatus = {
              type = "string"
            }
            verificationOutcome = {
              type = "string"
            }
          }
        }
      }
      pagination = {
        type = "object"
        properties = {
          total = {
            type = "integer"
          }
          limit = {
            type = "integer"
          }
          offset = {
            type = "integer"
          }
          nextOffset = {
            type = ["integer", "null"]
          }
        }
      }
    }
    required = ["results", "pagination"]
  })
}

resource "aws_api_gateway_model" "verification_lookup_result" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "VerificationLookupResult"
  description  = "Verification lookup result model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      results = {
        type = "array"
        items = {
          type = "object"
        }
      }
      pagination = {
        type = "object"
      }
    }
    required = ["results", "pagination"]
  })
}

resource "aws_api_gateway_model" "conversation_history" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "ConversationHistory"
  description  = "Conversation history model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      verificationId = {
        type = "string"
      }
      conversationAt = {
        type = "string"
      }
      currentTurn = {
        type = "integer"
      }
      maxTurns = {
        type = "integer"
      }
      turnStatus = {
        type = "string"
      }
      history = {
        type = "array"
        items = {
          type = "object"
        }
      }
    }
    required = ["verificationId", "conversationAt", "history"]
  })
}

resource "aws_api_gateway_model" "health_status" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "HealthStatus"
  description  = "Health status model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      status = {
        type = "string"
        enum = ["healthy", "degraded", "unhealthy"]
      }
      version = {
        type = "string"
      }
      timestamp = {
        type = "string"
      }
      services = {
        type = "object"
      }
    }
    required = ["status", "version", "timestamp", "services"]
  })
}

resource "aws_api_gateway_model" "presigned_url_response" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "PresignedUrlResponse"
  description  = "Presigned URL response model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      presignedUrl = {
        type = "string"
      }
      expiresAt = {
        type = "string"
      }
      contentType = {
        type = "string"
      }
    }
    required = ["presignedUrl", "expiresAt", "contentType"]
  })
}

resource "aws_api_gateway_model" "file_browser_response" {
  rest_api_id  = aws_api_gateway_rest_api.api.id
  name         = "FileBrowserResponse"
  description  = "File browser response model"
  content_type = "application/json"
  schema = jsonencode({
    type = "object"
    properties = {
      currentPath = {
        type = "string"
      }
      items = {
        type = "array"
        items = {
          type = "object"
          properties = {
            name = {
              type = "string"
            }
            type = {
              type = "string"
              enum = ["folder", "image"]
            }
            path = {
              type = "string"
            }
          }
        }
      }
    }
    required = ["currentPath", "items"]
  })
}
