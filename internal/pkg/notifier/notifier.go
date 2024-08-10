package notifier

import (
	"context"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/mezk/tg-reminder/internal/pkg/domain"
	"github.com/mezk/tg-reminder/internal/pkg/sender"
)

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

// Storage - storage interface.
type Storage interface {
	GetPendingReminders(ctx context.Context, limit int64) ([]domain.Reminder, error)
	UpdateReminder(ctx context.Context, reminder domain.Reminder) error
}

// BotResponseSender - bot's response sender.
type BotResponseSender interface {
	SendBotResponse(response sender.BotResponse, opts ...sender.BotResponseOption) error
}

// Notifier sends reminders to users.
type Notifier struct {
	botResponseSender BotResponseSender
	storage           Storage
	interval          time.Duration
}

// New creates new Notifier.
func New(responseSender BotResponseSender, storage Storage, interval time.Duration) *Notifier {
	return &Notifier{botResponseSender: responseSender, storage: storage, interval: interval}
}

// Run starts infinite loop to fetch reminders from Storage and send them to users.
// Breaks infinite loop on context error.
func (n *Notifier) Run(ctx context.Context) {
	log.Printf("[INFO] notifier started, tick interval %s", n.interval)

	ticker := time.NewTicker(n.interval)

	for {
		select {

		case <-ctx.Done():
			log.Printf("[INFO] notifier is shutting down")
			return

		case <-ticker.C:
			log.Printf("[DEBUG] notifier start sending reminders")

			const limit = 100
			reminders, err := n.storage.GetPendingReminders(ctx, limit)
			if err != nil {
				log.Printf("[ERROR] failed to fetch reminders: %v", err)
				continue
			}

			for _, r := range reminders {
				if err = n.botResponseSender.SendBotResponse(sender.BotResponse{
					ChatID: r.ChatID,
					Text:   r.FormatNotify(),
				}, sender.WithReminderDoneButton(r.ID)); err != nil {
					log.Printf("[ERROR] failed to send reminder %d: %v", r.ID, err)
				}

				log.Printf("[INFO] notifier sent reminder %d to user %d in chat %d", r.ID, r.UserID, r.ChatID)

				// delay reminder for 15 minutes to wait for ack from user
				r.RemindAt = timeNowUTC().Add(15 * time.Minute)
				r.AttemptsLeft--

				if r.AttemptsLeft == 0 {
					r.Status = domain.ReminderStatusAttemptsExhausted
				}

				if err = n.storage.UpdateReminder(ctx, r); err != nil {
					log.Printf("[ERROR] failed to update reminder %s: %v", r, err)
				}
			}
		}
	}
}
