package llm

import (
	"fmt"
	"strings"
)

// Email is the minimal input the LLM tasks operate on.
type Email struct {
	Subject    string
	Body       string
	Recipients []string
}

// styleSystemPrompt instructs the model to act as a corporate email style
// reviewer and to return a strict JSON array of suggestions.
const styleSystemPrompt = `You are ComplyMail, an assistant that reviews outbound corporate emails against a company style guide.
Analyze the email and return ONLY a JSON array of suggestion objects. Each object has:
  - "type": always "style"
  - "severity": one of "info", "warning", "error"
  - "message": a short, actionable recommendation
Return an empty array [] if the email fully complies. Do not include any prose outside the JSON.`

// sensitivitySystemPrompt instructs the model to classify how sensitive the
// email content is and to return a strict JSON object.
const sensitivitySystemPrompt = `You are ComplyMail, a data-loss-prevention classifier for outbound corporate emails.
Assess how sensitive the email content is (confidential data, internal-only material, credentials, personal data).
Return ONLY a JSON object with:
  - "level": one of "LOW", "MEDIUM", "HIGH"
  - "reasons": an array of short strings explaining the classification
Do not include any prose outside the JSON.`

// buildStyleMessages assembles the chat messages for a style analysis,
// injecting the company style guide and the email content.
func buildStyleMessages(styleGuide string, email Email) []Message {
	return []Message{
		{Role: RoleSystem, Content: styleSystemPrompt},
		{Role: RoleUser, Content: renderStyleUserPrompt(styleGuide, email)},
	}
}

// buildSensitivityMessages assembles the chat messages for a sensitivity
// classification.
func buildSensitivityMessages(email Email) []Message {
	return []Message{
		{Role: RoleSystem, Content: sensitivitySystemPrompt},
		{Role: RoleUser, Content: renderEmailBlock(email)},
	}
}

func renderStyleUserPrompt(styleGuide string, email Email) string {
	var b strings.Builder
	if strings.TrimSpace(styleGuide) != "" {
		b.WriteString("COMPANY STYLE GUIDE:\n")
		b.WriteString(styleGuide)
		b.WriteString("\n\n")
	}
	b.WriteString(renderEmailBlock(email))
	return b.String()
}

func renderEmailBlock(email Email) string {
	return fmt.Sprintf(
		"EMAIL\nRecipients: %s\nSubject: %s\n\nBody:\n%s",
		strings.Join(email.Recipients, ", "),
		email.Subject,
		email.Body,
	)
}
