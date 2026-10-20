package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v4"
)

const maxBatchSize = 100

type Sender struct {
	client *resend.Client
	from   string
}

func NewSender(apiKey, from string) *Sender {
	return &Sender{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

func (s *Sender) Send(ctx context.Context, msg Message, promotionID, category string, recipients []string) error {
	for start := 0; start < len(recipients); start += maxBatchSize {
		chunk := recipients[start:min(start+maxBatchSize, len(recipients))]

		params := make([]*resend.SendEmailRequest, 0, len(chunk))
		for _, to := range chunk {
			params = append(params, &resend.SendEmailRequest{
				From:    s.from,
				To:      []string{to},
				Subject: msg.Subject,
				Html:    msg.HTML,
				Text:    msg.Text,
				Tags: []resend.Tag{
					{Name: "type", Value: "promotion"},
					{Name: "category", Value: category},
				},
			})
		}

		res, err := s.client.Batch.SendWithOptions(ctx, params, &resend.BatchSendEmailOptions{
			IdempotencyKey:  fmt.Sprintf("batch-promotion/%s-%d", promotionID, start/maxBatchSize),
			BatchValidation: resend.BatchValidationPermissive,
		})
		if err != nil {
			return fmt.Errorf("failed to send promotion batch: %w", err)
		}

		for _, batchErr := range res.Errors {
			to := ""
			if batchErr.Index >= 0 && batchErr.Index < len(chunk) {
				to = chunk[batchErr.Index]
			}
			slog.Error("Failed to send promotion email",
				"promotionId", promotionID,
				"to", to,
				"error", batchErr.Message,
			)
		}

		slog.Info("Promotion emails sent",
			"promotionId", promotionID,
			"sent", len(res.Data),
			"failed", len(res.Errors),
		)
	}

	return nil
}
