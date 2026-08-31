package bird_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

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

// Sending to the sandbox bounce address, which hard-bounces every time. The
// tag and metadata are what make the resulting event findable in the logs.
func ExampleEmailService_Send_bounce() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:     "onboarding@messagebird.dev",
		To:       []string{"bounce+signup-flow@messagebird.dev"},
		Subject:  "Sandbox bounce test",
		HTML:     "<p>This message will hard-bounce.</p>",
		Tags:     []bird.Tag{{Name: "flow", Value: "signup"}},
		Metadata: map[string]any{"test_run": "docs-capture-1"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// A richer send: cc/bcc, reply-to, tags, metadata, opt-out of click tracking,
// and an idempotency key. The server deduplicates the request, so it is safe to retry.
// Send a published template in place of inline content. The template supplies
// the subject and bodies; Parameters fills its variables.
func ExampleEmailService_Send_template() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Send(context.Background(), bird.EmailSendParams{
		From:       "onboarding@messagebird.dev",
		To:         []string{"delivered@messagebird.dev"},
		Category:   "transactional",
		Template:   "welcome-email",
		Parameters: map[string]any{"first_name": "Jane"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

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
// failures (timeouts, 429, 5xx), so a returned error is terminal. Propagate it
// unless you need to act on a category.
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

// Cancel stops a message that has not left yet, including a scheduled send
// before its send time or a queued message still awaiting delivery.
func ExampleEmailService_Cancel() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Email.Cancel(context.Background(), "em_abc123"); err != nil {
		log.Fatal(err)
	}
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
func ExampleWebhooksService_Unwrap() {
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
func ExampleSmsService_Send() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Sms.Send(context.Background(), bird.SmsSendParams{
		From:     "+15557654321",
		To:       "+15551234567",
		Text:     "Your verification code is 123456.",
		Category: bird.SMSCategoryAuthentication,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// Send up to 100 independent messages in one call. Acceptance is
// all-or-nothing: every message is validated before any of them queue.
func ExampleSmsService_SendBatch() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	batch, err := client.Sms.SendBatch(context.Background(), bird.SmsSendBatchParams{
		Messages: []bird.SmsSendParams{
			{
				From: "+15557654321", To: "+15551111111",
				Text: "Hi Alice!", Category: bird.SMSCategoryMarketing,
			},
			{
				From: "+15557654321", To: "+15552222222",
				Text: "Hi Bob!", Category: bird.SMSCategoryMarketing,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, msg := range batch.Data {
		fmt.Println(msg.Id, *msg.Status)
	}
}

// Send an SMS from a stored template, supplying its variables.
func ExampleSmsService_Send_template() {
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

// List the SMS templates available to the workspace. The catalogue is small,
// returned in full, and not paginated.
func ExampleSmsTemplatesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	list, err := client.SmsTemplates.List(context.Background(), bird.SMSTemplateListParams{
		Scope: "system",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, tpl := range list.Data {
		fmt.Println(tpl.Id, *tpl.Slug)
	}
}

// Read one SMS template by its slug (or id).
func ExampleSmsTemplatesService_Get() {
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

// Send a WhatsApp template message.
func ExampleWhatsappService_Send() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	code := "123456"
	msg, err := client.Whatsapp.Send(context.Background(), bird.WhatsappSendParams{
		To:       "+15551234567",
		Template: "bird_otp",
		Language: "en",
		Components: []bird.WhatsAppMessageTemplateComponent{{
			Type:       "body",
			Parameters: &[]bird.WhatsAppMessageTemplateComponentParameter{{Type: "text", Text: &code}},
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// Send free-form WhatsApp text, inside an open 24-hour customer service window.
func ExampleWhatsappService_Send_freeForm() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Whatsapp.Send(context.Background(), bird.WhatsappSendParams{
		To:   "+15551234567",
		From: "+15557654321",
		Text: &bird.WhatsAppTextSend{Body: "Your order has shipped!"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, *msg.Status)
}

// Read a single WhatsApp message by id.
func ExampleWhatsappService_Get() {
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
func ExampleWhatsappService_List() {
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
func ExampleWhatsappService_ListEvents() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	events, err := client.Whatsapp.ListEvents(context.Background(), "wam_01krdgeqcxet5s7t44vh8rt9mg", bird.WhatsappListEventsParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range events.Data {
		fmt.Println(e.Id, e.Type)
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
	// Set the first name and clear the last name (Null sends an explicit JSON
	// null); omit a field to leave it unchanged.
	contact, err := client.Contacts.Update(context.Background(), "con_123", bird.ContactUpdateParams{
		FirstName: bird.Value("Jane"),
		LastName:  bird.Null[string](),
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
		Contacts: []bird.ContactCreateRequest{
			{Email: bird.Ptr(openapi_types.Email("a@x.com"))},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range result.Data {
		email := ""
		if item.Entry.Email != nil {
			email = *item.Entry.Email
		}
		fmt.Println(email, item.Status)
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
	// Rename the audience and clear its description (Null sends an explicit JSON
	// null). Omit a field to leave it unchanged; bird.Value(...) sets a new value.
	audience, err := client.Audiences.Update(context.Background(), "adn_123", bird.AudienceUpdateParams{
		Name:        "Renamed",
		Description: bird.Null[string](),
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
func ExampleVerifyVerificationsService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	verification, err := client.Verify.Verifications.Create(context.Background(), bird.VerifyVerificationsCreateParams{
		To: bird.VerificationTo{PhoneNumber: bird.String("+15551234567")},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(verification.Id, *verification.Status)
}

// Check the passcode a recipient submitted, identified by the same recipient.
func ExampleVerifyVerificationsService_Check() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Verify.Verifications.Check(context.Background(), bird.VerifyVerificationsCheckParams{
		To:   bird.VerificationTo{PhoneNumber: bird.String("+15551234567")},
		Code: "123456",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*result.Success)
}

// Send a fresh passcode on the next channel when the recipient never got the first.
func ExampleVerifyVerificationsService_NextChannel() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	verification, err := client.Verify.Verifications.NextChannel(context.Background(), bird.VerifyVerificationsNextChannelParams{
		To: bird.VerificationTo{PhoneNumber: bird.String("+15551234567")},
	})
	if err != nil {
		log.Fatal(err)
	}
	if verification.LastChannel != nil {
		fmt.Println(*verification.LastChannel)
	}
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

// BySendingIP ranks delivery statistics per sending IP.
func ExampleEmailStatsService_BySendingIP() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Stats.BySendingIP(context.Background(), bird.EmailStatsBySendingIPParams{
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
		Tracking: bird.Value(bird.DomainTrackingConfig{Name: "links"}),
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

// List auto-paginates across all mailboxes in the workspace.
func ExampleEmailMailboxesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for mailbox, err := range client.Email.Mailboxes.List(context.Background(), bird.EmailMailboxesListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(mailbox.Id)
	}
}

func ExampleEmailMailboxesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	mailbox, err := client.Email.Mailboxes.Create(context.Background(), bird.EmailMailboxesCreateParams{
		DisplayName: "Support",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mailbox.Id, *mailbox.Address)
}

func ExampleEmailMailboxesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	mailbox, err := client.Email.Mailboxes.Get(context.Background(), "mbx_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*mailbox.Address)
}

func ExampleEmailMailboxesService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	mailbox, err := client.Email.Mailboxes.Update(context.Background(), "mbx_123", bird.EmailMailboxesUpdateParams{
		DisplayName: bird.Value("Sales"),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mailbox.Id)
}

func ExampleEmailMailboxesService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Email.Mailboxes.Delete(context.Background(), "mbx_123"); err != nil {
		log.Fatal(err)
	}
}

func ExampleEmailMailboxesService_Restore() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	mailbox, err := client.Email.Mailboxes.Restore(context.Background(), "mbx_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mailbox.Id)
}

func ExampleEmailMailboxesService_Resume() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	mailbox, err := client.Email.Mailboxes.Resume(context.Background(), "mbx_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mailbox.Id)
}

func ExampleEmailMailboxesService_Stats() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Email.Mailboxes.Stats(context.Background(), "mbx_123", bird.EmailMailboxesStatsParams{})
	if err != nil {
		log.Fatal(err)
	}
	if stats.Summary != nil {
		if d := stats.Summary.Delivery; d != nil {
			fmt.Println(d.Delivered, d.Bounced)
		}
	}
}

func ExampleEmailMailboxesMessagesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Mailboxes.Messages.Create(context.Background(), "mbx_123", bird.EmailMailboxesMessagesCreateParams{
		To:      []string{"customer@example.com"},
		Subject: "Following up",
		HTML:    "<p>Hi, just checking in.</p>",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id)
}

func ExampleEmailMailboxesService_Labels() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	labels, err := client.Email.Mailboxes.Labels(context.Background(), "mbx_123")
	if err != nil {
		log.Fatal(err)
	}
	for _, l := range labels.Data {
		fmt.Println(l.Name)
	}
}

func ExampleEmailMailboxesReceiveRulesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for rule, err := range client.Email.Mailboxes.ReceiveRules.List(context.Background(), "mbx_123", bird.EmailMailboxesReceiveRulesListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(rule.Id, rule.Action, rule.Entry)
	}
}

func ExampleEmailMailboxesReceiveRulesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	rule, err := client.Email.Mailboxes.ReceiveRules.Create(context.Background(), "mbx_123", bird.EmailMailboxesReceiveRulesCreateParams{
		Action: "block",
		Entry:  "spam.example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rule.Id)
}

func ExampleEmailMailboxesReceiveRulesService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Email.Mailboxes.ReceiveRules.Delete(context.Background(), "mbx_123", "erl_456"); err != nil {
		log.Fatal(err)
	}
}

func ExampleEmailThreadsService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for thread, err := range client.Email.Threads.List(context.Background(), bird.EmailThreadsListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(thread.Id)
	}
}

func ExampleEmailThreadsService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	thread, err := client.Email.Threads.Get(context.Background(), "thr_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(thread.Id)
}

func ExampleEmailThreadsService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	thread, err := client.Email.Threads.Update(context.Background(), "thr_123", bird.EmailThreadsUpdateParams{
		Labels: &bird.EmailLabelsUpdate{Add: &[]string{"urgent"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(thread.Id)
}

func ExampleEmailThreadsService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Email.Threads.Delete(context.Background(), "thr_123", bird.EmailThreadsDeleteParams{Permanent: true}); err != nil {
		log.Fatal(err)
	}
}

func ExampleEmailThreadsMessagesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for msg, err := range client.Email.Threads.Messages.List(context.Background(), "thr_123", bird.EmailThreadsMessagesListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg.Id, msg.Direction)
	}
}

func ExampleEmailThreadsMessagesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	msg, err := client.Email.Threads.Messages.Get(context.Background(), "thr_123", "rem_456")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(msg.Id, msg.Direction)
}

func ExampleEmailThreadsMessagesService_Body() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	body, err := client.Email.Threads.Messages.Body(context.Background(), "thr_123", "rem_456")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(body.Text)
}

func ExampleEmailThreadsMessagesService_Reply() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	reply, err := client.Email.Threads.Messages.Reply(context.Background(), "thr_123", "rem_456", bird.EmailThreadsMessagesReplyParams{
		Text: "Thanks for reaching out!",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply.Id)
}

func ExampleEmailThreadsMessagesService_Attachments() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Email.Threads.Messages.Attachments(context.Background(), "thr_123", "rem_456")
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range result.Data {
		fmt.Println(a.Filename, a.Size)
	}
}

// Publish delivers one event to one or more channels. The Realtime app's own
// key and secret authenticate the call, alongside the workspace API key.
func ExampleRealtimeService_Publish() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Realtime.Publish(context.Background(), "rap_123", bird.RealtimePublishParams{
		Event:    "order.created",
		Channels: []string{"orders", "presence-lobby"},
		Data:     map[string]any{"order_id": "ord_1", "total": 4200},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Data)
}

// PublishBatch sends several events, each to a single channel, in one request.
func ExampleRealtimeService_PublishBatch() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Realtime.PublishBatch(context.Background(), "rap_123", bird.RealtimePublishBatchParams{
		Events: []bird.RealtimeBatchEventParams{
			{Event: "order.created", Channel: "orders", Data: map[string]any{"id": 1}},
			{Event: "order.updated", Channel: "orders", Data: map[string]any{"id": 2}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Data)
}

// AuthorizeChannel signs a subscription for a browser client locally. Authorize
// only after checking that the user may join the channel. The signature is the
// edge's only evidence that they may. On a private-encrypted channel, the
// response also carries that channel's shared_secret, so the client can decrypt.
func ExampleRealtimeService_AuthorizeChannel() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
		option.WithRealtimeEncryptionMasterKey(os.Getenv("BIRD_REALTIME_ENCRYPTION_MASTER_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/bird/auth", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ConnectionID string `json:"connection_id"`
			ChannelName  string `json:"channel_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		auth, err := client.Realtime.AuthorizeChannel(bird.RealtimeChannelAuthorizationParams{
			ConnectionID: req.ConnectionID,
			ChannelName:  req.ChannelName,
		})
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(auth)
	})
}

// List returns the app's occupied channels. The Realtime service does not
// paginate this listing, so one response holds every occupied channel.
func ExampleRealtimeChannelsService_List() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	channels, err := client.Realtime.Channels.List(context.Background(), "rap_123", bird.RealtimeChannelListParams{
		Prefix:  "presence-",
		Include: []bird.RealtimeChannelInclude{bird.RealtimeIncludeMemberCount},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, ch := range channels.Data {
		fmt.Println(ch.Name, ch.MemberCount)
	}
}

// Get reads one channel's occupancy, plus any counts named in Include.
func ExampleRealtimeChannelsService_Get() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	channel, err := client.Realtime.Channels.Get(context.Background(), "rap_123", "presence-lobby", bird.RealtimeChannelGetParams{
		Include: []bird.RealtimeChannelInclude{bird.RealtimeIncludeMemberCount},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(channel.Occupied, channel.MemberCount)
}

// Members lists the members present on a presence channel.
func ExampleRealtimeChannelsService_Members() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	members, err := client.Realtime.Channels.Members(context.Background(), "rap_123", "presence-lobby")
	if err != nil {
		log.Fatal(err)
	}
	for _, m := range members.Members {
		fmt.Println(m.MemberId)
	}
}

// Send delivers one event to every connection a single member holds, without
// putting it on a channel anyone else can subscribe to.
func ExampleRealtimeMembersService_Send() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Realtime.Members.Send(context.Background(), "rap_01krdgeqcxet5s7t44vh8rt9mg", "user_42", bird.RealtimeMemberSendParams{
		Event: "order-shipped",
		Data:  map[string]any{"order_id": "ord_123"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Disconnect closes every connection belonging to one member. Use it for a
// sign-out or ban flow.
func ExampleRealtimeMembersService_Disconnect() {
	client, err := bird.NewClient(
		option.WithAPIKey(os.Getenv("BIRD_API_KEY")),
		option.WithRealtimeCredentials(os.Getenv("BIRD_REALTIME_KEY"), os.Getenv("BIRD_REALTIME_SECRET")),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Realtime.Members.Disconnect(context.Background(), "rap_123", "member:42"); err != nil {
		log.Fatal(err)
	}
}

// List the workspace's calls. Filtering to the in-flight statuses gives the
// calls happening right now; omit the filter for completed records.
func ExampleVoiceService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for call, err := range client.Voice.List(context.Background(), bird.VoiceListParams{
		Status: []bird.VoiceCallStatus{"ringing", "in_progress"},
	}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(call.Id, call.Status)
	}
}

// Get returns one call at any point in its lifecycle, settled or still up.
func ExampleVoiceService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	call, err := client.Voice.Get(context.Background(), "vcl_01k0p3v9wera3v6q6xw3e9y2mh")
	if err != nil {
		log.Fatal(err)
	}
	// A call still ringing or connected carries no economics yet.
	fmt.Println(call.Status, call.DurationMs)
}

// Email tells you whether an address is worth sending to before you send.
func ExampleLookupService_Email() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	answer, err := client.Lookup.Email(context.Background(), bird.LookupEmailParams{
		Email: "aisha.khan@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	// result is an open vocabulary; delivery_confidence is always comparable.
	fmt.Println(*answer.Result, *answer.DeliveryConfidence)
}

// PhoneNumber returns the free baseline plus whichever paid blocks you ask for.
func ExampleLookupService_PhoneNumber() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	answer, err := client.Lookup.PhoneNumber(context.Background(), bird.LookupPhoneNumberParams{
		PhoneNumber: "+31612345678",
		Type:        []bird.LookupProperty{"classification", "score"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*answer.CountryCode, *answer.LineType)
	// Only a block whose status is ok carries a value, and only that one is billed.
	if answer.Score != nil && *answer.Score.Status == "ok" {
		fmt.Println(*answer.Score.Value)
	}
}

// ListEvents returns the lifecycle timeline for one message.
func ExampleSmsService_ListEvents() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	events, err := client.Sms.ListEvents(context.Background(), "sms_01j9x2k3m4n5p6q7r8s9t0v1w2", bird.SmsListEventsParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range events.Data {
		fmt.Println(*e.Type, *e.OccurredAt)
	}
}

// Summary returns the delivery and latency totals for a window.
func ExampleSmsStatsService_Summary() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	summary, err := client.Sms.Stats.Summary(context.Background(), bird.SmsStatsSummaryParams{
		From: "2026-05-01", // a calendar day for a day-grain window (up to 365 days), or
		To:   "2026-05-31", // an RFC 3339 instant (e.g. "2026-05-01T00:00:00Z") for hour-grain (up to 720 hours)
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*summary.Delivery.Accepted, *summary.Delivery.DeliveryRate)
}

// Daily returns one row per calendar day in the window.
func ExampleSmsStatsService_Daily() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	series, err := client.Sms.Stats.Daily(context.Background(), bird.SmsStatsDailyParams{
		From: time.Now().AddDate(0, 0, -7),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, point := range *series.Data {
		fmt.Println(*point.Bucket, *point.Delivery.Accepted)
	}
}

// Hourly returns one row per hour, for a window of at most 30 days.
func ExampleSmsStatsService_Hourly() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	series, err := client.Sms.Stats.Hourly(context.Background(), bird.SmsStatsHourlyParams{
		From: time.Now().Add(-24 * time.Hour),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, point := range *series.Data {
		fmt.Println(*point.Bucket, *point.Delivery.Accepted)
	}
}

// ByCountry ranks destination countries by the sort metric.
func ExampleSmsStatsService_ByCountry() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByCountry(context.Background(), bird.SmsStatsByCountryParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
		Sort: "delivery_rate", // worst delivery first is Sort plus a read of the tail
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Country, *row.Delivery.Accepted, *row.Delivery.DeliveryRate)
	}
}

// ByCarrier compares delivery across the carriers that handled the traffic.
func ExampleSmsStatsService_ByCarrier() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByCarrier(context.Background(), bird.SmsStatsByCarrierParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Carrier, *row.Delivery.Delivered)
	}
}

// ByCategory splits the window by the category messages were sent under.
func ExampleSmsStatsService_ByCategory() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByCategory(context.Background(), bird.SmsStatsByCategoryParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Category, *row.Delivery.Accepted)
	}
}

// ByOriginator compares how each sender address performs.
func ExampleSmsStatsService_ByOriginator() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByOriginator(context.Background(), bird.SmsStatsByOriginatorParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Originator, *row.Delivery.DeliveryRate)
	}
}

// ByStatus shows where the window's messages ended up.
func ExampleSmsStatsService_ByStatus() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByStatus(context.Background(), bird.SmsStatsByStatusParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Status, *row.Count)
	}
}

// ByErrorCode ranks the failure reasons behind undelivered traffic.
func ExampleSmsStatsService_ByErrorCode() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByErrorCode(context.Background(), bird.SmsStatsByErrorCodeParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		// The same value as the error_code filter on Sms.List, so a row joins to its messages.
		fmt.Println(*row.ErrorCode, *row.Delivery.Failed)
	}
}

// ByTag ranks the campaigns and segments sends are tagged with.
func ExampleSmsStatsService_ByTag() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.ByTag(context.Background(), bird.SmsStatsByTagParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		// A message carrying several tags counts once under each, so rows do not
		// sum to the period total.
		fmt.Println(*row.Tag, *row.Delivery.Accepted)
	}
}

// Summary returns how many messages the workspace's numbers received.
func ExampleSmsStatsInboundService_Summary() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	summary, err := client.Sms.Stats.Inbound.Summary(context.Background(), bird.SmsStatsInboundSummaryParams{
		From: "2026-05-01",
		To:   "2026-05-31",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*summary.Received)
}

// Daily returns received-message counts, one row per calendar day.
func ExampleSmsStatsInboundService_Daily() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	series, err := client.Sms.Stats.Inbound.Daily(context.Background(), bird.SmsStatsInboundDailyParams{
		From: time.Now().AddDate(0, 0, -7),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, point := range *series.Data {
		fmt.Println(*point.Bucket, *point.Received)
	}
}

// Hourly returns received-message counts, one row per hour.
func ExampleSmsStatsInboundService_Hourly() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	series, err := client.Sms.Stats.Inbound.Hourly(context.Background(), bird.SmsStatsInboundHourlyParams{
		From: time.Now().Add(-24 * time.Hour),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, point := range *series.Data {
		fmt.Println(*point.Bucket, *point.Received)
	}
}

// ByCountry groups received messages by where the sender messaged from.
func ExampleSmsStatsInboundService_ByCountry() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.Inbound.ByCountry(context.Background(), bird.SmsStatsInboundByCountryParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Country, *row.Received)
	}
}

// ByOperator groups received messages by the sender's mobile operator.
func ExampleSmsStatsInboundService_ByOperator() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.Inbound.ByOperator(context.Background(), bird.SmsStatsInboundByOperatorParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		// Messages whose operator the carrier did not report are excluded, so
		// these rows can sum to less than the inbound summary for the same period.
		fmt.Println(*row.MccMnc, *row.Received)
	}
}

// ByNumber shows which of the workspace's numbers took the traffic.
func ExampleSmsStatsInboundService_ByNumber() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	stats, err := client.Sms.Stats.Inbound.ByNumber(context.Background(), bird.SmsStatsInboundByNumberParams{
		From: time.Now().AddDate(0, -1, 0),
		To:   time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range *stats.Data {
		fmt.Println(*row.Number, *row.Received)
	}
}

// List iterates every suppression currently stopping the workspace's messages.
func ExampleSmsSuppressionsService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for suppression, err := range client.SmsSuppressions.List(context.Background(), bird.SmsSuppressionsListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(suppression.Originator, suppression.Destination, *suppression.Reason)
	}
}

// Get reads one suppression by id.
func ExampleSmsSuppressionsService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	suppression, err := client.SmsSuppressions.Get(context.Background(), "sup_01j9x2k3m4n5p6q7r8s9t0v1w2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(suppression.Originator, suppression.Destination, *suppression.Blocking)
}

// Add stops one sender from messaging one subscriber.
func ExampleSmsSuppressionsService_Add() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	// A suppression covers one sender and one subscriber, so stopping every
	// sender means one call per sender.
	suppression, err := client.SmsSuppressions.Add(context.Background(), bird.SmsSuppressionsAddParams{
		Destination: "+15550001234",
		Originator:  "+15557654321",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(suppression.Id)
}

// Remove ends a manual suppression.
func ExampleSmsSuppressionsService_Remove() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	// Only a `manual` suppression can be ended: a subscriber's own stop keyword
	// and a carrier's opt-out are refused.
	if err := client.SmsSuppressions.Remove(context.Background(), "sup_01j9x2k3m4n5p6q7r8s9t0v1w2"); err != nil {
		log.Fatal(err)
	}
}

// List shows what a reply to the workspace's numbers does.
func ExampleSmsKeywordRulesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	rules, err := client.SmsKeywordRules.List(context.Background(), bird.SmsKeywordRulesListParams{
		Country: "NL", // omit for every country Bird's catalogue covers, plus your own rules
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, rule := range rules.Data {
		fmt.Println(rule.Operation, rule.Keywords)
	}
}

// Get reads one rule, Bird's or the workspace's own.
func ExampleSmsKeywordRulesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	rule, err := client.SmsKeywordRules.Get(context.Background(), "skr_01j9x2k3m4n5p6q7r8s9t0v1w2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rule.Operation, *rule.Reply)
}

// Create overrides Bird's default reply for one country.
func ExampleSmsKeywordRulesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	rule, err := client.SmsKeywordRules.Create(context.Background(), bird.SmsKeywordRulesCreateParams{
		Operation: bird.SMSKeywordOperationStop,
		Country:   bird.Value("NL"),
		Reply:     bird.Value("You are unsubscribed from MyBrand. Reply START to resume."),
	})
	if err != nil {
		log.Fatal(err)
	}
	// EffectiveKeywords is Bird's set plus any of your own.
	fmt.Println(rule.Id, *rule.EffectiveKeywords)
}

// Update changes a rule the workspace created.
func ExampleSmsKeywordRulesService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	// Omitting Keywords leaves the set alone; an empty slice clears your
	// additions back to Bird's.
	rule, err := client.SmsKeywordRules.Update(context.Background(), "skr_01j9x2k3m4n5p6q7r8s9t0v1w2", bird.SmsKeywordRulesUpdateParams{
		Reply: bird.Value("You are unsubscribed. Reply START to resume."),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*rule.Reply)
}

// Delete restores Bird's default for that country and operation.
func ExampleSmsKeywordRulesService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.SmsKeywordRules.Delete(context.Background(), "skr_01j9x2k3m4n5p6q7r8s9t0v1w2"); err != nil {
		log.Fatal(err)
	}
}

// AvailableList finds numbers on sale in one country. The search is always
// country-scoped, so CountryCode is required.
func ExampleNumbersAvailableService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for candidate, err := range client.Numbers.Available.List(context.Background(), bird.NumbersAvailableListParams{
		CountryCode:  "GB",
		Capabilities: []string{"sms", "voice"},
	}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(candidate.Number, candidate.NumberType)
	}
}

// Get checks one number is still for sale. A number a carrier supplies is only
// on sale while the carrier still has it, so a 404 means someone else took it.
func ExampleNumbersAvailableService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	candidate, err := client.Numbers.Available.Get(context.Background(), "+447700900201")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(candidate.CountryCode)
}

// Create buys a number. Most orders finish inside the request.
func ExampleNumbersOrdersService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	order, err := client.Numbers.Orders.Create(context.Background(), bird.NumbersOrdersCreateParams{
		Number: "+447700900201",
	})
	if err != nil {
		log.Fatal(err)
	}
	// An order that has to wait on a carrier comes back without a NumberId.
	// Poll it until it is completed or failed.
	fmt.Println(order.Status, order.Id)
}

// Get polls an order that did not finish inline.
func ExampleNumbersOrdersService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	order, err := client.Numbers.Orders.Get(context.Background(), "nor_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		log.Fatal(err)
	}
	// FailureReason says what went wrong, and only ever on a failed order.
	fmt.Println(order.Status)
}

// List finds the purchases that did not complete.
func ExampleNumbersOrdersService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for order, err := range client.Numbers.Orders.List(context.Background(), bird.NumbersOrdersListParams{
		Status: "failed",
	}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(order.Number)
	}
}

// List walks the numbers allocated to the workspace.
func ExampleNumbersService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for allocated, err := range client.Numbers.List(context.Background(), bird.NumbersListParams{
		CountryCode: "GB",
	}) {
		if err != nil {
			log.Fatal(err)
		}
		// Kind distinguishes a number you bought from one Bird manages.
		fmt.Println(allocated.Number, allocated.Kind, allocated.Status)
	}
}

// Get reads one number allocated to the workspace.
func ExampleNumbersService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	allocated, err := client.Numbers.Get(context.Background(), "nda_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		log.Fatal(err)
	}
	// A country that asks for ownership paperwork answers on Ownership; most
	// answer nil.
	fmt.Println(allocated.Status)
}

// Release gives a dedicated number back, stopping its monthly charge.
func ExampleNumbersService_Release() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	// Only a dedicated number can be released; a shared one answers E14002.
	if err := client.Numbers.Release(context.Background(), "nda_01krdgeqcxet5s7t44vh8rt9mg"); err != nil {
		log.Fatal(err)
	}
}

// List looks up the messaging preferences recorded for one handle on one
// channel.
func ExamplePreferencesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for pref, err := range client.Preferences.List(context.Background(), bird.PreferencesListParams{
		Channel: bird.PreferenceChannelSms,
		Handle:  "+15550001234",
	}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(pref.Id, *pref.Status)
	}
}

// Get reads one recorded preference by ID.
func ExamplePreferencesService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	pref, err := client.Preferences.Get(context.Background(), "prf_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*pref.Status, *pref.Coverage)
}

// Create records a consent grant, with the evidence needed to override a
// stored opt-out on the same number.
func ExamplePreferencesService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Preferences.Create(context.Background(), bird.PreferencesCreateParams{
		Channel:     bird.PreferenceChannelEmail,
		Handle:      "recipient@example.com",
		Status:      bird.PreferenceStatusGranted,
		Source:      "signup-form-v2",
		ConsentedAt: time.Now(),
	})
	if err != nil {
		log.Fatal(err)
	}
	// A newer statement already on file answers Applied false instead of an
	// error, with the surviving statement in Preference.
	if result.Applied != nil && *result.Applied {
		fmt.Println("grant recorded")
	}
}

// Delete removes a preference statement by ID.
func ExamplePreferencesService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Preferences.Delete(context.Background(), "prf_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		log.Fatal(err)
	}
	// A newer statement recorded since refuses the delete: Applied comes back
	// false and Preference carries the statement that survived.
	switch {
	case result.Applied == nil:
	case !*result.Applied && result.Preference != nil:
		fmt.Println("delete refused, surviving statement:", result.Preference.Id)
	default:
		fmt.Println("deleted")
	}
}

// List reads the messaging preferences recorded for a contact's own handles
// across every channel.
func ExampleContactsPreferencesService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for pref, err := range client.Contacts.Preferences.List(context.Background(), "con_01krdgeqcxet5s7t44vh8rt9mg", bird.ContactsPreferencesListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(*pref.Channel, *pref.Status)
	}
}

func ExampleWorkspaceService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	workspace, err := client.Workspace.Get(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(workspace.Id, workspace.Name)
}

// Create returns the signing secret, which is never readable again.
func ExampleWebhooksService_Create() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	created, err := client.Webhooks.Create(context.Background(), bird.WebhooksCreateParams{
		URL:         "https://acme.com/hooks/bird",
		Events:      []bird.WebhookEventType{"email.delivered", "email.bounced"},
		Description: "Delivery pipeline",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(created.Id, created.Secret)
}

// List iterates every webhook endpoint in the workspace.
func ExampleWebhooksService_List() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	for endpoint, err := range client.Webhooks.List(context.Background(), bird.WebhooksListParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(endpoint.Id, endpoint.Url)
	}
}

// Get returns one endpoint's URL, subscribed events, and delivery status.
func ExampleWebhooksService_Get() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	endpoint, err := client.Webhooks.Get(context.Background(), "whk_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(endpoint.Url)
}

// Update replaces the whole subscription set rather than adding to it.
func ExampleWebhooksService_Update() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	endpoint, err := client.Webhooks.Update(context.Background(), "whk_123", bird.WebhooksUpdateParams{
		Events: []bird.WebhookEventType{"email.delivered"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(endpoint.Events)
}

// Test delivers a synthetic event to the endpoint.
func ExampleWebhooksService_Test() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Webhooks.Test(context.Background(), "whk_123", bird.WebhooksTestParams{
		EventType: "email.delivered",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Status)
}

// Attempts is one endpoint's delivery log.
func ExampleWebhooksService_Attempts() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	attempts, err := client.Webhooks.Attempts(context.Background(), "whk_123", bird.WebhooksAttemptsParams{})
	if err != nil {
		log.Fatal(err)
	}
	for _, attempt := range attempts.Data {
		fmt.Println(attempt.Status)
	}
}

// RotateSecret mints a new secret; both old and new sign deliveries for 24
// hours, after which the old one stops.
func ExampleWebhooksService_RotateSecret() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	rotated, err := client.Webhooks.RotateSecret(context.Background(), "whk_123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rotated.Secret)
}

// Delete stops delivery to the endpoint.
func ExampleWebhooksService_Delete() {
	client, err := bird.NewClient(option.WithAPIKey(os.Getenv("BIRD_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Webhooks.Delete(context.Background(), "whk_123"); err != nil {
		log.Fatal(err)
	}
}
