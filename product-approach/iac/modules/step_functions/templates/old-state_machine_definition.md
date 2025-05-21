{
  "Comment": "Kootoro GenAI Vending Machine Verification - Enhanced Two-Turn Workflow with Standardized Status Transitions",
  "StartAt": "InitializeVerificationContext",
  "States": {
    "InitializeVerificationContext": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion": "1.2.0",
        "verificationContext": {
          "verificationId.$": "States.UUID()",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "VERIFICATION_REQUESTED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId": "",
          "verificationAt.$": "$$.State.EnteredTime",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig": {
            "maxTurns": 2,
            "referenceImageTurn": 1,
            "checkingImageTurn": 2
          },
          "turnTimestamps": {
            "initialized.$": "$$.State.EnteredTime"
          },
          "requestMetadata": {
            "requestId.$": "States.UUID()",
            "requestTimestamp.$": "$$.State.EnteredTime",
            "processingStarted.$": "$$.State.EnteredTime"
          }
        }
      },
      "Next": "CheckPreviousVerificationId"
    },
    
    "CheckPreviousVerificationId": {
      "Type": "Choice",
      "Choices": [
        {
          "And": [
            {
              "Variable": "$.verificationContext.verificationType",
              "StringEquals": "PREVIOUS_VS_CURRENT"
            },
            {
              "Variable": "$.verificationContext.previousVerificationId",
              "IsPresent": true
            }
          ],
          "Next": "UpdatePreviousVerificationId"
        }
      ],
      "Default": "CheckVerificationType"
    },
    
    "UpdatePreviousVerificationId": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status.$": "$.verificationContext.status",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        }
      },
      "Next": "CheckVerificationType"
    },
    
    "CheckVerificationType": {
      "Type": "Choice",
      "Choices": [
        {
          "Variable": "$.verificationContext.verificationType",
          "StringEquals": "LAYOUT_VS_CHECKING",
          "Next": "SetStatusInitializing"
        },
        {
          "Variable": "$.verificationContext.verificationType",
          "StringEquals": "PREVIOUS_VS_CURRENT",
          "Next": "SetStatusInitializing"
        }
      ],
      "Default": "HandleInvalidVerificationType"
    },
    
    "SetStatusInitializing": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "VERIFICATION_INITIALIZED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        }
      },
      "Next": "InitializeVerification"
    },
    
    "InitializeVerification": {
      "Type": "Task",
      "Resource": "${function_arns["initialize"]}",
      "InputPath": "$",
      "ResultPath": "$.initializeResult",
      "Retry": [
        {
          "ErrorEquals": ["States.TaskFailed", "ServiceException", "ThrottlingException"],
          "IntervalSeconds": 2,
          "MaxAttempts": 3,
          "BackoffRate": 2.0
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "HandleInitializationError"
        }
      ],
      "Next": "CheckVerificationFlow"
    },
    
    "CheckVerificationFlow": {
      "Type": "Choice",
      "Choices": [
        {
          "Variable": "$.verificationContext.verificationType",
          "StringEquals": "PREVIOUS_VS_CURRENT",
          "Next": "FetchHistoricalVerification"
        }
      ],
      "Default": "FetchImages"
    },
    
    "HandleInvalidVerificationType": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "VERIFICATION_FAILED",
          "error": {
            "code": "INVALID_VERIFICATION_TYPE",
            "message": "Invalid verificationType. Must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT.",
            "timestamp.$": "$$.State.EnteredTime"
          }
        }
      },
      "End": true
    },
    
    "HandleInitializationError": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "INITIALIZATION_FAILED",
          "error.$": "$.error",
          "timestamp.$": "$$.State.EnteredTime"
        }
      },
      "End": true
    },
    
    "FetchHistoricalVerification": {
      "Type": "Task",
      "Resource": "${function_arns["fetch_historical_verification"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext"
      },
      "ResultPath": "$.historicalContext",
      "Retry": [
        {
          "ErrorEquals": ["States.TaskFailed", "ServiceException", "ThrottlingException"],
          "IntervalSeconds": 2,
          "MaxAttempts": 3,
          "BackoffRate": 2.0
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "HandleHistoricalFetchError"
        }
      ],
      "Next": "SetStatusFetchingImages"
    },
    
    "HandleHistoricalFetchError": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "HISTORICAL_FETCH_FAILED",
          "error.$": "$.error",
          "timestamp.$": "$$.State.EnteredTime"
        }
      },
      "End": true
    },
    
    "SetStatusFetchingImages": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "FETCHING_IMAGES",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext"
      },
      "Next": "FetchImages"
    },
    
    "FetchImages": {
      "Type": "Task",
      "Resource": "${function_arns["fetch_images"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext"
      },
      "ResultPath": "$.fetchImagesResult",
      "Retry": [
        {
          "ErrorEquals": ["States.TaskFailed", "ServiceException", "ThrottlingException"],
          "IntervalSeconds": 2,
          "MaxAttempts": 3,
          "BackoffRate": 2.0
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "HandleFetchImagesError"
        }
      ],
      "Next": "SetStatusImagesFetched"
    },
    
    "HandleFetchImagesError": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "IMAGE_FETCH_FAILED",
          "error.$": "$.error",
          "timestamp.$": "$$.State.EnteredTime"
        }
      },
      "End": true
    },
    
    "SetStatusImagesFetched": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "IMAGES_FETCHED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps": {
            "initialized.$": "$.verificationContext.turnTimestamps.initialized",
            "imagesFetched.$": "$$.State.EnteredTime"
          },
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext": {
          "$": "States.JsonMerge(States.StringToJson('{}'), States.JsonMerge($.historicalContext || {}, {}))"
        },
        "images.$": "$.fetchImagesResult.images",
        "layoutMetadata.$": "$.fetchImagesResult.layoutMetadata"
      },
      "Next": "PrepareSystemPrompt"
    },
    
    "PrepareSystemPrompt": {
      "Type": "Task",
      "Resource": "${function_arns["prepare_system_prompt"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "images.$": "$.images",
        "layoutMetadata.$": "$.layoutMetadata",
        "historicalContext.$": "$.historicalContext"
      },
      "ResultPath": "$.systemPrompt",
      "Next": "SetStatusPromptPrepared"
    },
    
    "SetStatusPromptPrepared": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "PROMPT_PREPARED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState": {
          "currentTurn": 0,
          "maxTurns": 2,
          "history": [],
          "referenceAnalysis": {},
          "checkingAnalysis": {}
        }
      },
      "Next": "PrepareTurn1Prompt"
    },
    
    "PrepareTurn1Prompt": {
      "Type": "Task",
      "Resource": "${function_arns["prepare_turn1_prompt"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "historicalContext.$": "$.historicalContext",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "turnNumber": 1,
        "includeImage": "reference"
      },
      "ResultPath": "$.currentPrompt",
      "Next": "SetStatusTurn1PromptReady"
    },
    
    "SetStatusTurn1PromptReady": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "TURN1_PROMPT_READY",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps": {
            "initialized.$": "$.verificationContext.turnTimestamps.initialized",
            "imagesFetched.$": "$.verificationContext.turnTimestamps.imagesFetched",
            "turn1Started.$": "$$.State.EnteredTime"
          },
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "currentPrompt.$": "$.currentPrompt"
      },
      "Next": "ExecuteTurn1"
    },
    
    "ExecuteTurn1": {
      "Type": "Task",
      "Resource": "${function_arns["execute_turn1"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "currentPrompt.$": "$.currentPrompt",
        "conversationState.$": "$.conversationState",
        "historicalContext.$": "$.historicalContext",
        "layoutMetadata.$": "$.layoutMetadata",
        "bedrockConfig": {
          "anthropic_version.$": "$.systemPrompt.bedrockConfig.anthropic_version",
          "max_tokens.$": "$.systemPrompt.bedrockConfig.max_tokens",
          "thinking": {
            "type": "enabled",
            "budget_tokens.$": "$.systemPrompt.bedrockConfig.thinking.budget_tokens"
          }
        }
      },
      "ResultPath": "$.turn1Response",
      "Retry": [
        {
          "ErrorEquals": ["ServiceException", "ThrottlingException"],
          "IntervalSeconds": 3,
          "MaxAttempts": 5,
          "BackoffRate": 2.0
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "HandleBedrockError"
        }
      ],
      "Next": "SetStatusTurn1Completed"
    },
    
    "SetStatusTurn1Completed": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "TURN1_COMPLETED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps": {
            "initialized.$": "$.verificationContext.turnTimestamps.initialized",
            "imagesFetched.$": "$.verificationContext.turnTimestamps.imagesFetched",
            "turn1Started.$": "$.verificationContext.turnTimestamps.turn1Started",
            "turn1Completed.$": "$$.State.EnteredTime"
          },
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "currentPrompt.$": "$.currentPrompt",
        "turn1Response.$": "$.turn1Response"
      },
      "Next": "ProcessTurn1Response"
    },
    
    "ProcessTurn1Response": {
      "Type": "Task",
      "Resource": "${function_arns["process_turn1_response"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "turn1Response.$": "$.turn1Response",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "currentPrompt.$": "$.currentPrompt",
        "conversationState.$": "$.conversationState",
        "layoutMetadata.$": "$.layoutMetadata",
        "historicalContext.$": "$.historicalContext"
      },
      "ResultPath": "$.referenceAnalysis",
      "Next": "SetStatusTurn1Processed"
    },
    
    "SetStatusTurn1Processed": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "TURN1_PROCESSED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "currentPrompt.$": "$.currentPrompt",
        "turn1Response.$": "$.turn1Response",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "conversationState": {
          "currentTurn": 1,
          "maxTurns": 2,
          "history.$": "States.Array($.conversationState.history, $.turn1Response)",
          "referenceAnalysis.$": "$.referenceAnalysis",
          "checkingAnalysis.$": "$.conversationState.checkingAnalysis"
        }
      },
      "Next": "PrepareTurn2Prompt"
    },
    
    "PrepareTurn2Prompt": {
      "Type": "Task",
      "Resource": "${function_arns["prepare_turn2_prompt"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "historicalContext.$": "$.historicalContext",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "turnNumber": 2,
        "includeImage": "checking",
        "previousContext.$": "$.referenceAnalysis"
      },
      "ResultPath": "$.currentPrompt",
      "Next": "SetStatusTurn2PromptReady"
    },
    
    "SetStatusTurn2PromptReady": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "TURN2_PROMPT_READY",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps": {
            "initialized.$": "$.verificationContext.turnTimestamps.initialized",
            "imagesFetched.$": "$.verificationContext.turnTimestamps.imagesFetched",
            "turn1Started.$": "$.verificationContext.turnTimestamps.turn1Started",
            "turn1Completed.$": "$.verificationContext.turnTimestamps.turn1Completed",
            "turn2Started.$": "$$.State.EnteredTime"
          },
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "currentPrompt.$": "$.currentPrompt",
        "turn1Response.$": "$.turn1Response",
        "referenceAnalysis.$": "$.referenceAnalysis"
      },
      "Next": "ExecuteTurn2"
    },
    
    "ExecuteTurn2": {
      "Type": "Task",
      "Resource": "${function_arns["execute_turn2"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "currentPrompt.$": "$.currentPrompt",
        "conversationState.$": "$.conversationState",
        "historicalContext.$": "$.historicalContext",
        "layoutMetadata.$": "$.layoutMetadata",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "turn1Response.$": "$.turn1Response"
      },
      "ResultPath": "$.turn2Response",
      "Retry": [
        {
          "ErrorEquals": ["ServiceException", "ThrottlingException"],
          "IntervalSeconds": 3,
          "MaxAttempts": 5,
          "BackoffRate": 2.0
        }
      ],
      "Catch": [
        {
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "HandleBedrockError"
        }
      ],
      "Next": "SetStatusTurn2Completed"
    },
    
    "HandleBedrockError": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "BEDROCK_PROCESSING_FAILED",
          "error.$": "$.error",
          "timestamp.$": "$$.State.EnteredTime"
        },
        "conversationState.$": "$.conversationState",
        "images.$": "$.images"
      },
      "Next": "FinalizeWithError"
    },
    
    "FinalizeWithError": {
      "Type": "Task",
      "Resource": "${function_arns["finalize_with_error"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "error.$": "$.error",
        "conversationState.$": "$.conversationState",
        "images.$": "$.images"
      },
      "ResultPath": "$.finalResults",
      "Next": "SetStatusWithError"
    },
    
    "SetStatusWithError": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "VERIFICATION_FAILED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "error.$": "$.error"
        },
        "finalResults.$": "$.finalResults"
      },
      "Next": "StoreResults"
    },
    
    "SetStatusTurn2Completed": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "TURN2_COMPLETED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps": {
            "initialized.$": "$.verificationContext.turnTimestamps.initialized",
            "imagesFetched.$": "$.verificationContext.turnTimestamps.imagesFetched",
            "turn1Started.$": "$.verificationContext.turnTimestamps.turn1Started",
            "turn1Completed.$": "$.verificationContext.turnTimestamps.turn1Completed",
            "turn2Started.$": "$.verificationContext.turnTimestamps.turn2Started",
            "turn2Completed.$": "$$.State.EnteredTime"
          },
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "currentPrompt.$": "$.currentPrompt",
        "turn1Response.$": "$.turn1Response",
        "turn2Response.$": "$.turn2Response",
        "referenceAnalysis.$": "$.referenceAnalysis"
      },
      "Next": "ProcessTurn2Response"
    },
    
    "ProcessTurn2Response": {
      "Type": "Task",
      "Resource": "${function_arns["process_turn2_response"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "turn2Response.$": "$.turn2Response",
        "turn1Response.$": "$.turn1Response",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "conversationState.$": "$.conversationState",
        "layoutMetadata.$": "$.layoutMetadata",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "currentPrompt.$": "$.currentPrompt",
        "historicalContext.$": "$.historicalContext"
      },
      "ResultPath": "$.checkingAnalysis",
      "Next": "SetStatusTurn2Processed"
    },
    
    "SetStatusTurn2Processed": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "TURN2_PROCESSED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "currentPrompt.$": "$.currentPrompt",
        "turn1Response.$": "$.turn1Response",
        "turn2Response.$": "$.turn2Response",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "checkingAnalysis.$": "$.checkingAnalysis",
        "conversationState": {
          "currentTurn": 2,
          "maxTurns": 2,
          "history.$": "States.Array($.conversationState.history, $.turn2Response)",
          "referenceAnalysis.$": "$.referenceAnalysis",
          "checkingAnalysis.$": "$.checkingAnalysis"
        }
      },
      "Next": "FinalizeResults"
    },
    
    "FinalizeResults": {
      "Type": "Task",
      "Resource": "${function_arns["finalize_results"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "conversationState.$": "$.conversationState",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "checkingAnalysis.$": "$.checkingAnalysis",
        "images.$": "$.images",
        "layoutMetadata.$": "$.layoutMetadata",
        "historicalContext.$": "$.historicalContext"
      },
      "ResultPath": "$.finalResults",
      "Next": "SetStatusResultsFinalized"
    },
    
    "SetStatusResultsFinalized": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "RESULTS_FINALIZED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "checkingAnalysis.$": "$.checkingAnalysis",
        "finalResults.$": "$.finalResults"
      },
      "Next": "StoreResults"
    },
    
    "StoreResults": {
      "Type": "Task",
      "Resource": "${function_arns["store_results"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "finalResults.$": "$.finalResults",
        "conversationState.$": "$.conversationState",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "checkingAnalysis.$": "$.checkingAnalysis",
        "images.$": "$.images",
        "layoutMetadata.$": "$.layoutMetadata",
        "historicalContext.$": "$.historicalContext"
      },
      "ResultPath": "$.storageResult",
      "Next": "SetStatusResultsStored"
    },
    
    "SetStatusResultsStored": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "RESULTS_STORED",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "historicalContext.$": "$.historicalContext",
        "images.$": "$.images",
        "systemPrompt.$": "$.systemPrompt",
        "layoutMetadata.$": "$.layoutMetadata",
        "conversationState.$": "$.conversationState",
        "referenceAnalysis.$": "$.referenceAnalysis",
        "checkingAnalysis.$": "$.checkingAnalysis",
        "finalResults.$": "$.finalResults",
        "storageResult.$": "$.storageResult"
      },
      "Next": "ShouldNotify"
    },
    
    "ShouldNotify": {
      "Type": "Choice",
      "Choices": [
        {
          "Variable": "$.verificationContext.notificationEnabled",
          "BooleanEquals": true,
          "Next": "Notify"
        }
      ],
      "Default": "SetStatusCompleted"
    },
    
    "Notify": {
      "Type": "Task",
      "Resource": "${function_arns["notify"]}",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext.$": "$.verificationContext",
        "finalResults.$": "$.finalResults",
        "storageResult.$": "$.storageResult"
      },
      "ResultPath": "$.notificationResult",
      "Next": "SetStatusNotificationSent"
    },
    
    "SetStatusNotificationSent": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationContext": {
          "verificationId.$": "$.verificationContext.verificationId",
          "verificationType.$": "$.verificationContext.verificationType",
          "status": "NOTIFICATION_SENT",
          "referenceImageUrl.$": "$.verificationContext.referenceImageUrl",
          "checkingImageUrl.$": "$.verificationContext.checkingImageUrl",
          "vendingMachineId.$": "$.verificationContext.vendingMachineId",
          "layoutId.$": "$.verificationContext.layoutId",
          "layoutPrefix.$": "$.verificationContext.layoutPrefix",
          "previousVerificationId.$": "$.verificationContext.previousVerificationId",
          "verificationAt.$": "$.verificationContext.verificationAt",
          "notificationEnabled.$": "$.verificationContext.notificationEnabled",
          "turnConfig.$": "$.verificationContext.turnConfig",
          "turnTimestamps.$": "$.verificationContext.turnTimestamps",
          "requestMetadata.$": "$.verificationContext.requestMetadata"
        },
        "finalResults.$": "$.finalResults",
        "storageResult.$": "$.storageResult",
        "notificationResult.$": "$.notificationResult"
      },
      "Next": "SetStatusCompleted"
    },
    
    "SetStatusCompleted": {
      "Type": "Pass",
      "Parameters": {
        "schemaVersion.$": "$.schemaVersion",
        "verificationId.$": "$.verificationContext.verificationId",
        "verificationType.$": "$.verificationContext.verificationType",
        "status": "COMPLETED",
        "timestamp.$": "$$.State.EnteredTime",
        "result": {
          "verificationStatus.$": "$.finalResults.verificationStatus",
          "resultImageUrl.$": "$.storageResult.resultImageUrl",
          "confidenceScore.$": "$.finalResults.confidenceScore",
          "discrepanciesCount.$": "$.finalResults.discrepanciesCount"
        },
        "turnTimestamps": {
          "initialized.$": "$.verificationContext.turnTimestamps.initialized",
          "imagesFetched.$": "$.verificationContext.turnTimestamps.imagesFetched",
          "turn1Started.$": "$.verificationContext.turnTimestamps.turn1Started",
          "turn1Completed.$": "$.verificationContext.turnTimestamps.turn1Completed",
          "turn2Started.$": "$.verificationContext.turnTimestamps.turn2Started",
          "turn2Completed.$": "$.verificationContext.turnTimestamps.turn2Completed",
          "completed.$": "$$.State.EnteredTime"
        }
      },
      "End": true
    }
  }
}