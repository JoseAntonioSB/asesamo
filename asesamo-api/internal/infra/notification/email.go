package notification

import (
	"context"
	"fmt"
	"sync"
)

// EmailSender defines the contract for sending transactional emails.
type EmailSender interface {
	SendEmail(ctx context.Context, to string, subject string, htmlBody string) error
}

// SentEmailRecord holds history of sent emails for testing and auditing.
type SentEmailRecord struct {
	To       string
	Subject  string
	HTMLBody string
}

// MockEmailSender is an in-memory thread-safe email sender for development and testing.
type MockEmailSender struct {
	mu      sync.RWMutex
	history []SentEmailRecord
}

// NewMockEmailSender constructs an initialized MockEmailSender.
func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		history: make([]SentEmailRecord, 0),
	}
}

// SendEmail records the email in history.
func (m *MockEmailSender) SendEmail(ctx context.Context, to string, subject string, htmlBody string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = append(m.history, SentEmailRecord{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
	})
	return nil
}

// GetSentEmails returns a copy of sent email history.
func (m *MockEmailSender) GetSentEmails() []SentEmailRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]SentEmailRecord, len(m.history))
	copy(res, m.history)
	return res
}

// FormatDocumentRejectionEmail generates official Mexican Spanish notification for rejected contractor compliance documents.
func FormatDocumentRejectionEmail(companyName, docType, workerName, reason string) (subject string, htmlBody string) {
	subject = fmt.Sprintf("ASESAMO: Documento Rechazado — %s (%s)", docType, companyName)
	workerSection := ""
	if workerName != "" {
		workerSection = fmt.Sprintf("<p><strong>Trabajador asignado:</strong> %s</p>", workerName)
	}

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #09090b; color: #f4f4f5; padding: 24px;">
  <div style="max-width: 600px; margin: 0 auto; background-color: #18181b; border: 1px solid #27272a; border-radius: 12px; padding: 32px;">
    <div style="font-family: monospace; font-size: 14px; font-weight: bold; color: #10b981; margin-bottom: 8px;">ASESAMO &bull; Notificación EHS</div>
    <h2 style="color: #ef4444; margin-top: 0;">Observación en Validación Documental</h2>
    <p>Estimado equipo de <strong>%s</strong>:</p>
    <p>Le informamos que el documento <strong>%s</strong> ha sido revisado por el departamento de Seguridad e Higiene Industrial y ha sido <strong>RECHAZADO</strong> debido a inconsistencias normativas.</p>
    %s
    <div style="background-color: #27272a; border-left: 4px solid #ef4444; padding: 16px; margin: 20px 0; border-radius: 4px;">
      <div style="font-size: 12px; font-weight: bold; color: #a1a1aa; text-transform: uppercase;">Motivo del Rechazo:</div>
      <div style="font-size: 14px; color: #f4f4f5; margin-top: 4px;">%s</div>
    </div>
    <p style="font-size: 13px; color: #a1a1aa;">Por favor ingrese al portal de contratistas de ASESAMO para corregir y cargar el expediente nuevamente antes del ingreso programado a planta.</p>
    <hr style="border: 0; border-top: 1px solid #27272a; margin: 24px 0;" />
    <p style="font-size: 11px; color: #71717a; text-align: center;">ASESAMO Industrial Platform &bull; Cumplimiento STPS Digital &bull; No responda a este correo automatizado.</p>
  </div>
</body>
</html>`, companyName, docType, workerSection, reason)

	return subject, htmlBody
}

// FormatDocumentExpirationWarningEmail generates early warning notice for documents expiring in 7, 15, or 30 days.
func FormatDocumentExpirationWarningEmail(companyName, docType string, daysRemaining int) (subject string, htmlBody string) {
	subject = fmt.Sprintf("ASESAMO: Aviso de Vencimiento Próximo en %d días — %s", daysRemaining, docType)

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #09090b; color: #f4f4f5; padding: 24px;">
  <div style="max-width: 600px; margin: 0 auto; background-color: #18181b; border: 1px solid #27272a; border-radius: 12px; padding: 32px;">
    <div style="font-family: monospace; font-size: 14px; font-weight: bold; color: #10b981; margin-bottom: 8px;">ASESAMO &bull; Alerta Preventiva de Vigencia</div>
    <h2 style="color: #f59e0b; margin-top: 0;">Documentación Próxima a Expirar</h2>
    <p>Estimado equipo de <strong>%s</strong>:</p>
    <p>Le notificamos que el documento <strong>%s</strong> vencerá en <strong>%d días</strong> naturales.</p>
    <div style="background-color: #27272a; border-left: 4px solid #f59e0b; padding: 16px; margin: 20px 0; border-radius: 4px;">
      <p style="margin: 0; font-size: 13px; color: #f4f4f5;">Para evitar bloqueos automáticos en torniquetes y casetas de acceso de planta, le sugerimos cargar la renovación con anticipación en la plataforma ASESAMO.</p>
    </div>
    <hr style="border: 0; border-top: 1px solid #27272a; margin: 24px 0;" />
    <p style="font-size: 11px; color: #71717a; text-align: center;">ASESAMO Industrial Platform &bull; Candado Inteligente de Acceso</p>
  </div>
</body>
</html>`, companyName, docType, daysRemaining)

	return subject, htmlBody
}

// FormatCriticalIncidentEmail generates high-priority notification for plant EHS staff.
func FormatCriticalIncidentEmail(incidentNumber, severity, location, description string) (subject string, htmlBody string) {
	subject = fmt.Sprintf("¡ALERTA EHS ASESAMO! Evento Reportado: %s (Severidad: %s)", incidentNumber, severity)

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #09090b; color: #f4f4f5; padding: 24px;">
  <div style="max-width: 600px; margin: 0 auto; background-color: #18181b; border: 2px solid #ef4444; border-radius: 12px; padding: 32px;">
    <div style="font-family: monospace; font-size: 14px; font-weight: bold; color: #ef4444; margin-bottom: 8px;">EMERGENCIA / INCIDENTE OPERATIVO EN PLANTA</div>
    <h2 style="color: #ffffff; margin-top: 0;">Reporte de Evento Folio %s</h2>
    <p><strong>Ubicación:</strong> %s</p>
    <p><strong>Nivel de Severidad:</strong> <span style="color: #ef4444; text-transform: uppercase; font-weight: bold;">%s</span></p>
    <div style="background-color: #27272a; padding: 16px; margin: 20px 0; border-radius: 4px;">
      <div style="font-size: 12px; font-weight: bold; color: #a1a1aa; text-transform: uppercase;">Descripción de los Hechos:</div>
      <div style="font-size: 14px; color: #f4f4f5; margin-top: 4px;">%s</div>
    </div>
    <p style="font-size: 13px; color: #a1a1aa;">El equipo médico y las brigadas EHS deben presentarse de inmediato en el área señalada según el protocolo NOM-021-STPS.</p>
  </div>
</body>
</html>`, incidentNumber, location, severity, description)

	return subject, htmlBody
}
