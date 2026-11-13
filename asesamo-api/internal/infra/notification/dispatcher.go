package notification

import (
	"context"
	"log"
)

// Dispatcher defines high-level notification dispatch methods for operational EHS and compliance events.
type Dispatcher interface {
	NotifyDocumentRejection(ctx context.Context, recipientEmail string, companyName string, docType string, workerName string, reason string) error
	NotifyDocumentExpiring(ctx context.Context, recipientEmail string, companyName string, docType string, daysRemaining int) error
	NotifyCriticalIncident(ctx context.Context, incidentNumber string, severity string, location string, description string) error
}

// Service coordinates email and webhook dispatches.
type Service struct {
	emailSender   EmailSender
	webhookSender WebhookSender
	webhookURL    string
	webhookSecret string
}

// NewService constructs an initialized notification Service.
func NewService(email EmailSender, webhook WebhookSender, webhookURL, webhookSecret string) *Service {
	return &Service{
		emailSender:   email,
		webhookSender: webhook,
		webhookURL:    webhookURL,
		webhookSecret: webhookSecret,
	}
}

// NotifyDocumentRejection sends an email alert to the contractor and fires an outgoing webhook.
func (s *Service) NotifyDocumentRejection(ctx context.Context, recipientEmail, companyName, docType, workerName, reason string) error {
	if s.emailSender != nil && recipientEmail != "" {
		subj, body := FormatDocumentRejectionEmail(companyName, docType, workerName, reason)
		if err := s.emailSender.SendEmail(ctx, recipientEmail, subj, body); err != nil {
			log.Printf("[Notification] Error sending rejection email to %s: %v", recipientEmail, err)
		}
	}

	if s.webhookSender != nil && s.webhookURL != "" {
		payload := map[string]interface{}{
			"event":        "document.rejected",
			"company_name": companyName,
			"doc_type":     docType,
			"worker_name":  workerName,
			"reason":       reason,
		}
		if err := s.webhookSender.SendWebhook(ctx, s.webhookURL, s.webhookSecret, "document.rejected", payload); err != nil {
			log.Printf("[Notification] Error sending rejection webhook: %v", err)
		}
	}

	return nil
}

// NotifyDocumentExpiring dispatches proactive expiration warnings.
func (s *Service) NotifyDocumentExpiring(ctx context.Context, recipientEmail, companyName, docType string, daysRemaining int) error {
	if s.emailSender != nil && recipientEmail != "" {
		subj, body := FormatDocumentExpirationWarningEmail(companyName, docType, daysRemaining)
		if err := s.emailSender.SendEmail(ctx, recipientEmail, subj, body); err != nil {
			log.Printf("[Notification] Error sending expiration email to %s: %v", recipientEmail, err)
		}
	}

	if s.webhookSender != nil && s.webhookURL != "" {
		payload := map[string]interface{}{
			"event":          "document.expiring",
			"company_name":   companyName,
			"doc_type":       docType,
			"days_remaining": daysRemaining,
		}
		if err := s.webhookSender.SendWebhook(ctx, s.webhookURL, s.webhookSecret, "document.expiring", payload); err != nil {
			log.Printf("[Notification] Error sending expiration webhook: %v", err)
		}
	}

	return nil
}

// NotifyCriticalIncident fires immediate alerts to plant emergency responders and webhooks.
func (s *Service) NotifyCriticalIncident(ctx context.Context, incidentNumber, severity, location, description string) error {
	if s.emailSender != nil {
		subj, body := FormatCriticalIncidentEmail(incidentNumber, severity, location, description)
		_ = s.emailSender.SendEmail(ctx, "ehs-alerts@asesamo.internal", subj, body)
	}

	if s.webhookSender != nil && s.webhookURL != "" {
		payload := map[string]interface{}{
			"event":           "incident.critical",
			"incident_number": incidentNumber,
			"severity":        severity,
			"location":        location,
			"description":     description,
		}
		if err := s.webhookSender.SendWebhook(ctx, s.webhookURL, s.webhookSecret, "incident.critical", payload); err != nil {
			log.Printf("[Notification] Error sending incident webhook: %v", err)
		}
	}

	return nil
}
