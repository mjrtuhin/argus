package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackSender struct {
	webhookURL string
	httpClient *http.Client
}

func NewSlackSender(webhookURL string) *SlackSender {
	return &SlackSender{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SlackSender) SendAlert(ctx context.Context, metricName string, value float64, score float64, severity string) error {
	// Build alert message
	emoji := getSeverityEmoji(severity)
	color := getSeverityColor(severity)
	
	message := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"blocks": []map[string]interface{}{
					{
						"type": "header",
						"text": map[string]string{
							"type": "plain_text",
							"text": fmt.Sprintf("%s Anomaly Detected!", emoji),
						},
					},
					{
						"type": "section",
						"fields": []map[string]string{
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Metric:*\n%s", metricName),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Severity:*\n%s", severity),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Value:*\n%.2f", value),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Score:*\n%.3f", score),
							},
						},
					},
					{
						"type": "context",
						"elements": []map[string]string{
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("Detected at %s", time.Now().Format("2006-01-02 15:04:05")),
							},
						},
					},
				},
			},
		},
	}

	// If no webhook URL, just log
	if s.webhookURL == "" {
		fmt.Printf("📢 [SLACK ALERT] %s %s anomaly detected: %s = %.2f (score: %.3f)\n",
			emoji, severity, metricName, value, score)
		return nil
	}

	// Send to Slack
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🚨"
	case "high":
		return "⚠️"
	case "medium":
		return "⚡"
	default:
		return "ℹ️"
	}
}

func getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#ff0000"
	case "high":
		return "#ff6b00"
	case "medium":
		return "#ffcc00"
	default:
		return "#36a64f"
	}
}
