package bird

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/apierror"
	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// WhatsappMessagesService reaches the subresources of one WhatsApp message. The
// channel's own message verbs stay on Client.Whatsapp; reach this via
// Client.Whatsapp.Messages.
type WhatsappMessagesService struct{ resource }

// WhatsAppMedia is media downloaded from a received WhatsApp message.
// ContentType is what storage declared, which is the message's own mime_type.
type WhatsAppMedia struct {
	Data          []byte
	ContentType   string
	ContentLength int64
}

// Media downloads the media on a received WhatsApp message — an image, video,
// audio clip, sticker or document. mediaID is the id on the message's content
// object, which Client.Whatsapp.Get returns.
//
// Media is kept for 30 days after the message arrives; after that the message
// still lists the media's mime_type and caption, and this returns a 410
// *APIError. Outbound messages carry no stored media.
//
// The second leg runs on http.DefaultClient rather than the client's own:
// option.WithHTTPClient may carry auth middleware, and the presigned target
// refuses a second auth mechanism.
func (s *WhatsappMessagesService) Media(ctx context.Context, messageID, mediaID string, opts ...option.RequestOption) (*WhatsAppMedia, error) {
	body, meta, err := s.getRedirect(ctx, opts, http.StatusFound, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapiNoRedirect.GetWhatsAppMessageMedia(ctx, oapi.WhatsAppMessageID(messageID), oapi.WhatsAppFileID(mediaID), cfg...)
	})
	if err != nil {
		return nil, err
	}
	// A 2xx is an edge answering with the bytes directly, which is also the only
	// arm the conformance corpus can script — its responses carry no headers, so
	// no 302 with a Location.
	if meta.Status != http.StatusFound {
		return &WhatsAppMedia{
			Data:          body,
			ContentType:   mediaContentType(meta.Header.Get("Content-Type")),
			ContentLength: int64(len(body)),
		}, nil
	}

	location := meta.Header.Get("Location")
	if location == "" {
		return nil, &apierror.ConnectionError{Err: errors.New("media redirect carried no Location header")}
	}
	return fetchStoredMedia(ctx, location)
}

// fetchStoredMedia takes the second leg with a bare client: the presigned URL
// carries its own credential and refuses a second auth mechanism, so nothing
// from this SDK's request path may ride along. Its failures surface as
// ConnectionError rather than going through apierror.FromResponse — a storage
// XML body is no Bird error envelope, and a 403 mapped that way would report
// the caller's own key as lacking permission.
func fetchStoredMedia(ctx context.Context, location string) (*WhatsAppMedia, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, &apierror.ConnectionError{Err: err}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &apierror.ConnectionError{Err: fmt.Errorf("downloading media: %w; call Media again for a fresh link", err)}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		return nil, &apierror.ConnectionError{Err: fmt.Errorf(
			"storage refused the download link (status %d): the link expired or was refused, call Media again for a fresh link",
			resp.StatusCode)}
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, &apierror.ConnectionError{Err: err}
	}
	return &WhatsAppMedia{
		Data:          data,
		ContentType:   mediaContentType(resp.Header.Get("Content-Type")),
		ContentLength: int64(len(data)),
	}, nil
}

func mediaContentType(value string) string {
	if value == "" {
		return "application/octet-stream"
	}
	return value
}
