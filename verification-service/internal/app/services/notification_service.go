package services

import (
	"context"
	"fmt"

	"verification-service/internal/domain/models"
)

// NotificationConfig contains notification configuration
type NotificationConfig struct {
	SNSTopicARN string
	WebhookURL  string
}

// NotificationService sends notifications about verification results
type NotificationService struct {
	config NotificationConfig
}

// NewNotificationService creates a new notification service
func NewNotificationService(config NotificationConfig) *NotificationService {
	return &NotificationService{
		config: config,
	}
}

// SendNotifications sends notifications about a verification
func (s *NotificationService) SendNotifications(
	ctx context.Context,
	verificationContext *models.VerificationContext,
	results *models.VerificationResult,
	resultImageURL string,
) error {
	// Check if notifications are configured
	if s.config.SNSTopicARN == "" && s.config.WebhookURL == "" {
		return nil
	}
	
	// Prepare notification message
	message := fmt.Sprintf("Verification completed for vending machine %s.\n", results.VendingMachineID)
	
	if results.VerificationStatus == models.StatusIncorrect {
		message += fmt.Sprintf("ALERT: %d discrepancies found!\n", results.VerificationSummary.DiscrepantPositions)
		message += fmt.Sprintf("Summary: %s\n", results.VerificationSummary.VerificationOutcome)
	} else {
		message += "All products correctly placed. No discrepancies found.\n"
	}
	
	message += fmt.Sprintf("\nAccuracy: %.1f%%\n", results.VerificationSummary.OverallAccuracy)
	message += fmt.Sprintf("View results: %s", resultImageURL)
	
	// Send SNS notification if configured
	if s.config.SNSTopicARN != "" {
		if err := s.sendSNSNotification(ctx, results.VerificationID, message); err != nil {
			return fmt.Errorf("failed to send SNS notification: %w", err)
		}
	}
	
	// Send webhook notification if configured
	if s.config.WebhookURL != "" {
		if err := s.sendWebhookNotification(ctx, results, message); err != nil {
			return fmt.Errorf("failed to send webhook notification: %w", err)
		}
	}
	
	return nil
}

// sendSNSNotification sends a notification via SNS
func (s *NotificationService) sendSNSNotification(ctx context.Context, verificationID, message string) error {
	// In a real implementation, this would use the AWS SDK to send an SNS notification
	// For this example, we'll just log the notification
	fmt.Printf("SNS Notification for %s: %s\n", verificationID, message)
	return nil
}

// sendWebhookNotification sends a notification via webhook
func (s *NotificationService) sendWebhookNotification(ctx context.Context, results *models.VerificationResult, message string) error {
	// In a real implementation, this would make an HTTP request to the webhook URL
	// For this example, we'll just log the notification
	fmt.Printf("Webhook Notification: %s\n", message)
	return nil
}