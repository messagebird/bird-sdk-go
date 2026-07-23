package bird_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

// Example constructs a client and sends an email. The region is taken from the
// API key's prefix; pass option.WithBaseURL or option.WithRegion to override.
func Example() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "onboarding@messagebird.dev",
		To:      []string{"delivered@messagebird.dev"},
		Subject: "Hello from Bird",
		HTML:    "<p>My first Bird email.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id)
}

func ExampleEmailService_Send() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "onboarding@messagebird.dev",
		To:      []string{"delivered@messagebird.dev"},
		Subject: "Hello from Bird",
		HTML:    "<p>My first Bird email.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// A richer send: cc/bcc, reply-to, tags, metadata, opt-out of click tracking,
// and an idempotency key (safe to retry — the server dedupes).
func ExampleEmailService_Send_rich() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Email.Send(context.Background(), bird.EmailSendParams{
		From:        "hello@acme.com",
		To:          []string{"a@example.com", "b@example.com"},
		Cc:          []string{"manager@example.com"},
		ReplyTo:     []string{"support@acme.com"},
		Subject:     "Your March invoice",
		HTML:        "<p>Attached.</p>",
		Tags:        []bird.EmailTag{{Name: "category", Value: "billing"}},
		Metadata:    map[string]any{"invoice_id": "inv_123"},
		TrackClicks: bird.Bool(false),
	}, option.WithIdempotencyKey("invoice-march/cust_1"))
	if err != nil {
		log.Fatal(err)
	}
}

// Send with display names: "Name <addr>" syntax in From and To.
func ExampleEmailService_Send_displayNames() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Email.Send(context.Background(), bird.EmailSendParams{
		From:    "Bird Support <support@acme.com>",
		To:      []string{"Jane Doe <jane@example.com>", "bob@example.com"},
		Subject: "Your order is confirmed",
		HTML:    "<p>Thanks for your order!</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
}

// SendBatch queues several emails in one request and returns one result item
// per message, in submission order.
func ExampleEmailService_SendBatch() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	batch, err := client.Email.SendBatch(context.Background(), bird.EmailSendBatchParams{
		Messages: []bird.EmailSendParams{
			{
				From:    "onboarding@messagebird.dev",
				To:      []string{"alice@example.com"},
				Subject: "Hello, Alice",
				HTML:    "<p>Welcome!</p>",
			},
			{
				From:    "onboarding@messagebird.dev",
				To:      []string{"bob@example.com"},
				Subject: "Hello, Bob",
				HTML:    "<p>Welcome!</p>",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range batch.Data {
		fmt.Println(item.Id)
	}
}

// Branch on the typed error hierarchy. The SDK already retries transient
// failures (timeouts, 429, 5xx), so a returned error is terminal — most callers
// just propagate it; branch only to act on a category.
func ExampleEmailService_Send_errors() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Email.Send(context.Background(), bird.EmailSendParams{
		From: "onboarding@messagebird.dev", To: []string{"delivered@messagebird.dev"}, Subject: "Hello from Bird", HTML: "<p>My first Bird email.</p>",
	})
	if err != nil {
		var rle *bird.RateLimitError
		var ve *bird.ValidationError
		var ae *bird.APIError
		switch {
		case errors.As(err, &rle):
			fmt.Println("rate limited; retry after", rle.RetryAfter)
		case errors.As(err, &ve):
			for _, d := range ve.Details {
				fmt.Printf("%s: %s\n", d.Param, d.Message)
			}
		case errors.As(err, &ae):
			fmt.Printf("API error %s (status %d, request %s)\n", ae.Code, ae.StatusCode, ae.RequestID)
		default:
			log.Print(err) // transport: *bird.ConnectionError or *bird.TimeoutError
		}
	}
}

func ExampleEmailService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Get(context.Background(), "em_abc123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*msg.Status, *msg.DeliveredCount)
}

// List auto-paginates: it lazily fetches each page and yields every matching
// message across all of them.
func ExampleEmailService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for msg, err := range client.Email.List(context.Background(), bird.EmailListParams{Status: bird.EmailStatusBounced}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg.Id)
	}
	page, err := client.Email.ListPage(context.Background(), bird.EmailListParams{}, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(page.Data)) // page.NextCursor carries the next starting_after
}

// EmailDefaults set common send fields once; a per-send value always wins.
func ExampleEmailDefaults() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithEmailDefaults(bird.EmailDefaults{
			From:     "hello@acme.com",
			Category: bird.CategoryTransactional,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	// From is filled from the default.
	if _, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		To: []string{"customer@example.com"}, Subject: "Hi", HTML: "<p>hi</p>",
	}); err != nil {
		log.Fatal(err)
	}
}

// Unwrap verifies the Standard Webhooks signature over the raw request body and
// returns a typed event to dispatch on.
func ExampleWebhookService_Unwrap() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithWebhookSecret(os.Getenv("BIRD_WEBHOOK_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/webhooks/bird", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		event, err := client.Webhooks.Unwrap(body, r.Header)
		if err != nil {
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent) // ack fast, then process

		payload, _ := event.AsAny()
		switch p := payload.(type) {
		case bird.EmailDeliveredEvent:
			fmt.Println("delivered:", p.Data.EmailId, p.Data.Recipient)
		case bird.EmailBouncedEvent:
			fmt.Println("bounced:", p.Type)
		}
	})
}

// The verb methods reach endpoints outside the curated surface, decoding the
// response into a value you provide.
func ExampleClient_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	var out struct {
		Data []struct {
			Recipient string `json:"recipient"`
		} `json:"data"`
	}
	if err := client.Get(context.Background(), "/v1/email/suppressions", &out); err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(out.Data))
}

// Send a free-text SMS.
func ExampleSMSService_Send() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Sms.Send(context.Background(), bird.SmsSendParams{
		To:       "+15551234567",
		Text:     "Your verification code is 123456.",
		Category: bird.SMSCategoryAuthentication,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// Send an SMS from a stored template, supplying its variables.
func ExampleSMSService_Send_template() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Sms.Send(context.Background(), bird.SmsSendParams{
		To:         "+15551234567",
		Template:   "bird_otp_verification",
		Parameters: map[string]any{"code": "123456"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id)
}

// List the SMS templates available to the workspace. The catalogue is small and
// returned in full — this list is not paginated.
func ExampleSMSTemplatesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	list, err := client.SmsTemplates.List(context.Background(), bird.SMSTemplateListParams{
		Scope: bird.SMSTemplateScopeSystem,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, tpl := range list.Data {
		fmt.Println(tpl.Id, *tpl.Name)
	}
}

// Read one SMS template by its name (or id).
func ExampleSMSTemplatesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	tpl, err := client.SmsTemplates.Get(context.Background(), "bird_otp_verification")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tpl.Id, *tpl.Body)
}

// Send a WhatsApp template message. Templates are currently the only supported
// content type.
func ExampleWhatsAppService_Send() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Whatsapp.Send(context.Background(), bird.WhatsappSendParams{
		To:       "+15551234567",
		Template: "bird_otp",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// Read a single WhatsApp message by id.
func ExampleWhatsAppService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Whatsapp.Get(context.Background(), "wam_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// List WhatsApp messages to a given contact, paginating lazily.
func ExampleWhatsAppService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for msg, err := range client.Whatsapp.List(context.Background(), bird.WhatsappListParams{PhoneNumber: "+15551234567"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg.Id)
	}
}

// List the lifecycle events for a WhatsApp message, in chronological order.
func ExampleWhatsAppService_ListEvents() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	events, err := client.Whatsapp.ListEvents(context.Background(), "wam_01krdgeqcxet5s7t44vh8rt9mg", bird.WhatsappListEventsParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range events.Data {
		fmt.Println(e.Id, *e.Type)
	}
}

// List the WhatsApp templates available to the workspace. The catalogue is
// small and returned in full — this list is not paginated.
func ExampleWhatsAppTemplatesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	list, err := client.WhatsappTemplates.List(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, tpl := range list.Data {
		fmt.Println(*tpl.Name)
	}
}

// Create a contact. Unset optional fields are omitted from the request.
func ExampleContactsService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	contact, err := client.Contacts.Create(context.Background(), bird.ContactCreateParams{
		Email:     "jane@acme.com",
		FirstName: "Jane",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contact.Id)
}

// Get returns a single contact by id.
func ExampleContactsService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	contact, err := client.Contacts.Get(context.Background(), "con_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contact.Email)
}

// List auto-paginates: it lazily fetches each page and yields every matching
// contact across all of them.
func ExampleContactsService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for contact, err := range client.Contacts.List(context.Background(), bird.ContactListParams{Limit: 50}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(contact.Id, contact.Email)
	}
}

// Update changes only the fields set in params; every other field is left
// unchanged.
func ExampleContactsService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	contact, err := client.Contacts.Update(context.Background(), "con_123", bird.ContactUpdateParams{
		FirstName: bird.String("Jane"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contact.Id)
}

// Delete removes a contact.
func ExampleContactsService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Contacts.Delete(context.Background(), "con_123"); err != nil {
		log.Fatal(err)
	}
}

// Batch creates or updates several contacts, matched by email address, in one
// request.
func ExampleContactsService_Batch() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Contacts.Batch(context.Background(), bird.ContactBatchParams{
		Contacts: []bird.ContactCreateParams{
			{Email: "a@x.com"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range result.Data {
		fmt.Println(item.Email, item.Status)
	}
}

func ExampleAudiencesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	audience, err := client.Audiences.Create(context.Background(), bird.AudienceCreateParams{
		Name: "Newsletter subscribers",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(audience.Id)
}

// Get returns a single audience by id.
func ExampleAudiencesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	audience, err := client.Audiences.Get(context.Background(), "adn_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(audience.Name)
}

// List auto-paginates: it lazily fetches each page and yields every matching
// audience across all of them.
func ExampleAudiencesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for audience, err := range client.Audiences.List(context.Background(), bird.AudienceListParams{Limit: 50}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(audience.Id, audience.Name)
	}
}

// Update changes only the fields set in params; every other field is left
// unchanged.
func ExampleAudiencesService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	audience, err := client.Audiences.Update(context.Background(), "adn_123", bird.AudienceUpdateParams{
		Name: bird.String("Renamed"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(audience.Id)
}

// Delete removes an audience. The contacts themselves are not deleted.
func ExampleAudiencesService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Audiences.Delete(context.Background(), "adn_123"); err != nil {
		log.Fatal(err)
	}
}

// ListContacts auto-paginates: it lazily fetches each page and yields every
// member of the audience across all of them.
func ExampleAudiencesService_ListContacts() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for member, err := range client.Audiences.ListContacts(context.Background(), "adn_123", bird.AudienceListContactsParams{Limit: 50}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(member.Contact.Id, member.Contact.Email)
	}
}

// AddContacts adds up to 1,000 existing contacts to a static audience.
func ExampleAudiencesService_AddContacts() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Audiences.AddContacts(context.Background(), "adn_123", bird.AudienceAddContactsParams{
		ContactIDs: []string{"con_1", "con_2"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// RemoveContacts removes up to 1,000 contacts from a static audience. The
// contacts themselves are not deleted.
func ExampleAudiencesService_RemoveContacts() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Audiences.RemoveContacts(context.Background(), "adn_123", bird.AudienceRemoveContactsParams{
		ContactIDs: []string{"con_1", "con_2"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// RemoveContact removes one contact's membership in an audience. The contact
// itself is not deleted and remains a member of any other audiences.
func ExampleAudiencesService_RemoveContact() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Audiences.RemoveContact(context.Background(), "adn_123", "con_1"); err != nil {
		log.Fatal(err)
	}
}

func ExampleContactPropertiesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	property, err := client.ContactProperties.Create(context.Background(), bird.ContactPropertyCreateParams{
		Key:  "plan",
		Type: "string",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(property.Id)
}

// Get returns a single contact property by id.
func ExampleContactPropertiesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	property, err := client.ContactProperties.Get(context.Background(), "prp_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(property.Key)
}

// List auto-paginates: it lazily fetches each page and yields every matching
// contact property across all of them.
func ExampleContactPropertiesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for property, err := range client.ContactProperties.List(context.Background(), bird.ContactPropertyListParams{Limit: 50}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(property.Id, property.Key)
	}
}

// Update changes a contact property's fallback value.
func ExampleContactPropertiesService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	property, err := client.ContactProperties.Update(context.Background(), "prp_123", bird.ContactPropertyUpdateParams{
		FallbackValue: "free",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(property.Id)
}

// Archive archives a contact property: the key stops being accepted in new
// contact writes, but every value already stored on contacts is preserved.
func ExampleContactPropertiesService_Archive() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	property, err := client.ContactProperties.Archive(context.Background(), "prp_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(property.Archived)
}

// Unarchive reactivates an archived contact property.
func ExampleContactPropertiesService_Unarchive() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	property, err := client.ContactProperties.Unarchive(context.Background(), "prp_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(property.Archived)
}

// Start a verification: send a one-time passcode over SMS.
func ExampleVerificationsService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	verification, err := client.Verify.Verifications.Create(context.Background(), bird.VerificationCreateParams{
		Phone: "+15551234567",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(verification.Id, *verification.Status)
}

// Check the passcode a recipient submitted, identified by the same recipient.
func ExampleVerificationsService_Check() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Verify.Verifications.Check(context.Background(), bird.VerificationCheckParams{
		Phone: "+15551234567",
		Code:  "123456",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*result.Success)
}

// Summary returns the delivery, engagement, and latency totals for a window.
func ExampleEmailStatsService_Summary() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	summary, err := client.Email.Stats.Summary(context.Background(), bird.EmailStatsSummaryParams{
		From: "2026-05-01", // a calendar day for a day-grain window (up to 365 days), or
		To:   "2026-05-31", // an RFC 3339 instant (e.g. "2026-05-01T00:00:00Z") for hour-grain (up to 720 hours)
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(summary.SendsAccepted)
}

// Daily returns one row per calendar day in the window.
func ExampleEmailStatsService_Daily() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	series, err := client.Email.Stats.Daily(context.Background(), bird.EmailStatsDailyParams{
		From: time.Now().AddDate(0, 0, -7),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(series.Data)
}

// Hourly returns one row per hour in the window (max 720 hours).
func ExampleEmailStatsService_Hourly() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	series, err := client.Email.Stats.Hourly(context.Background(), bird.EmailStatsHourlyParams{
		From: time.Now().Add(-24 * time.Hour),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(series.Data)
}

// ByTag ranks statistics per tag.
func ExampleEmailStatsService_ByTag() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByTag(context.Background(), bird.EmailStatsByTagParams{
		Sort:  "opens",
		Limit: 10,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByCategory ranks statistics per category.
func ExampleEmailStatsService_ByCategory() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByCategory(context.Background(), bird.EmailStatsByCategoryParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// BySendingIp ranks delivery statistics per sending IP.
func ExampleEmailStatsService_BySendingIp() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.BySendingIp(context.Background(), bird.EmailStatsBySendingIpParams{
		Sort: "bounced",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// BySendingDomain ranks statistics per sending domain.
func ExampleEmailStatsService_BySendingDomain() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.BySendingDomain(context.Background(), bird.EmailStatsBySendingDomainParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByRecipientDomain ranks statistics per recipient mailbox domain.
func ExampleEmailStatsService_ByRecipientDomain() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByRecipientDomain(context.Background(), bird.EmailStatsByRecipientDomainParams{
		Limit: 20,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByMailboxProvider ranks post-delivery statistics per mailbox provider.
func ExampleEmailStatsService_ByMailboxProvider() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByMailboxProvider(context.Background(), bird.EmailStatsByMailboxProviderParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByMailboxProviderRegion ranks post-delivery statistics per provider region.
func ExampleEmailStatsService_ByMailboxProviderRegion() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByMailboxProviderRegion(context.Background(), bird.EmailStatsByMailboxProviderRegionParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByTemplate ranks statistics per template.
func ExampleEmailStatsService_ByTemplate() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByTemplate(context.Background(), bird.EmailStatsByTemplateParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByLocation ranks engagement statistics per geographic location.
func ExampleEmailStatsService_ByLocation() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByLocation(context.Background(), bird.EmailStatsByLocationParams{
		GroupBy: "country",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByClient ranks engagement statistics per reading environment.
func ExampleEmailStatsService_ByClient() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByClient(context.Background(), bird.EmailStatsByClientParams{
		GroupBy: "email_client",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByBounceCode ranks bounce counts per SMTP error code.
func ExampleEmailStatsService_ByBounceCode() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByBounceCode(context.Background(), bird.EmailStatsByBounceCodeParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByComplaintType ranks complaint counts per complaint type.
func ExampleEmailStatsService_ByComplaintType() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByComplaintType(context.Background(), bird.EmailStatsByComplaintTypeParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// ByBroadcast ranks statistics per broadcast.
func ExampleEmailStatsService_ByBroadcast() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.ByBroadcast(context.Background(), bird.EmailStatsByBroadcastParams{
		Limit: 25,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stats.Data)
}

// Register a sending domain. It returns in "pending" with the DNS records to
// publish; call Verify once they are in place.
func ExampleDomainsService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	domain, err := client.Domains.Create(context.Background(), bird.DomainCreateParams{
		Domain: "mail.acme.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(domain.Id, *domain.Status)
}

// Get returns a single sending domain by id, with its DNS records.
func ExampleDomainsService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	domain, err := client.Domains.Get(context.Background(), "dom_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*domain.Domain)
}

// List auto-paginates: it lazily fetches each page and yields every sending
// domain across all of them.
func ExampleDomainsService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for domain, err := range client.Domains.List(context.Background(), bird.DomainListParams{Limit: 50}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(domain.Id, *domain.Status)
	}
}

// Update edits a sending domain. Only the fields you set change.
func ExampleDomainsService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	domain, err := client.Domains.Update(context.Background(), "dom_123", bird.DomainUpdateParams{
		Settings: &bird.DomainSettings{ClickTracking: bird.Bool(true), OpenTracking: bird.Bool(true)},
		Tracking: &bird.DomainTrackingConfig{Name: "links"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(domain.Id)
}

// Delete removes a sending domain.
func ExampleDomainsService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Domains.Delete(context.Background(), "dom_123"); err != nil {
		log.Fatal(err)
	}
}

// Verify triggers a fresh DNS check and returns the refreshed domain. Safe to
// repeat while waiting for DNS to propagate.
func ExampleDomainsService_Verify() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	domain, err := client.Domains.Verify(context.Background(), "dom_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*domain.Status)
}
