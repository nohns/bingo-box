package mail

import (
	"context"
	"fmt"
	"io"

	"github.com/mailgun/mailgun-go/v4"
	bingo "github.com/nohns/bingo-box/server"
)

type Mailer struct {
	client *mailgun.MailgunImpl

	BaseDownloadLink string
}

func (m *Mailer) SendCards(ctx context.Context, to *bingo.Player, cardsFile io.ReadCloser) error {
	return m.sendCardsWithMailgun(ctx, to, cardsFile)
}

func (m *Mailer) sendCardsWithMailgun(ctx context.Context, to *bingo.Player, cardsFile io.ReadCloser) error {
	dlLink := fmt.Sprintf("%s/player/%s/downloadCards", m.BaseDownloadLink, to.ID)
	sender := "Bingo Box <info@bingobox.io>"
	subject := "Your bingo cards are ready"
	body := fmt.Sprintf(`
	<html>
		<body>
			<h1>Here. Your cards are ready!</h1>
			<p>
				Print them out by opening the pdf file attached to this email or get them by clicking 
				<a href="%s">here</a>
			</p>
			<p>
				Best regards,<br/>
				Bingo box
			</p>
		</body>
	</html>
	`, dlLink)
	recipient := to.Email

	msg := m.client.NewMessage(sender, subject, "", recipient)
	msg.SetHtml(body)
	msg.SetReplyTo("Bingo box support <contact@bingobox.io>")
	msg.AddReaderAttachment("cards.pdf", cardsFile)

	_, _, err := m.client.Send(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func NewMailer(domain, apiKey string) *Mailer {
	mg := mailgun.NewMailgun(domain, apiKey)

	return &Mailer{
		client: mg,
	}
}
