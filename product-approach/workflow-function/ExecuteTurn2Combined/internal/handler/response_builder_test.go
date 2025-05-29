package handler

import (
	"testing"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
)

func TestBuildTurn2StepFunctionResponseRefs(t *testing.T) {
	builder := NewResponseBuilder(config.Config{})
	req := &models.Turn2Request{
		VerificationID: "verif-1",
		S3Refs: models.Turn2RequestS3Refs{
			Prompts: models.PromptRefs{System: models.S3Reference{Bucket: "b", Key: "sys"}},
			Images:  models.Turn2ImageRefs{CheckingBase64: models.S3Reference{Bucket: "b", Key: "img"}},
			Turn1: models.Turn1References{
				RawResponse:       models.S3Reference{Bucket: "b", Key: "t1raw"},
				ProcessedResponse: models.S3Reference{Bucket: "b", Key: "t1proc"},
			},
		},
	}
	resp := &models.Turn2Response{
		S3Refs: models.Turn2ResponseS3Refs{
			RawResponse:       models.S3Reference{Bucket: "b", Key: "raw"},
			ProcessedResponse: models.S3Reference{Bucket: "b", Key: "proc"},
		},
		Status: models.StatusTurn2Completed,
	}

	out := builder.BuildTurn2StepFunctionResponse(req, resp)
	responses := out.S3References["responses"].(map[string]interface{})
	if responses["turn2Raw"].(models.S3Reference).Key != "raw" {
		t.Fatalf("missing turn2Raw")
	}
	if responses["turn2Processed"].(models.S3Reference).Key != "proc" {
		t.Fatalf("missing turn2Processed")
	}
}
