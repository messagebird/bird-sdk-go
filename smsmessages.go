package bird

import (
	"context"
	"errors"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/apierror"
	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// SmsMessagesService reaches the subresources of one SMS message. The
// channel's own message verbs stay on Client.Sms; reach this via
// Client.Sms.Messages.
type SmsMessagesService struct{ resource }

// SMSMediaDownload is media downloaded from a received MMS message.
// ContentType is what storage declared, which is the message's own mime_type.
type SMSMediaDownload struct {
	Data          []byte
	ContentType   string
	ContentLength int64
}

// Media downloads the media on a received MMS message: an image, video,
// audio clip or document. mediaID is the id on the message's media
// entry, which Client.Sms.Get returns.
//
// Media expires with its message, 30 days after it arrives, and this returns a
// 404 *APIError from then on. A 410 *APIError means only the file expired: the
// message still lists the media's mime_type and filename. Messages sent through
// Bird carry no stored media.
//
// The second leg runs on a bare client rather than the client's own:
// option.WithHTTPClient may carry auth middleware, and the presigned target
// refuses a second auth mechanism. It is one attempt within the request's
// timeout: option.WithMaxRetries retries the API leg only, so a failed or
// timed-out download is retried by calling this again, which mints a fresh link.
func (s *SmsMessagesService) Media(ctx context.Context, messageID, mediaID string, opts ...option.RequestOption) (*SMSMediaDownload, error) {
	body, meta, timeout, err := s.getRedirect(ctx, opts, http.StatusFound, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapiNoRedirect.GetSMSMessageMedia(ctx, oapi.SMSMessageID(messageID), oapi.SMSFileID(mediaID), nil, cfg...)
	})
	if err != nil {
		return nil, err
	}
	// A 2xx is an edge answering with the bytes directly, which is also the only
	// arm the conformance corpus can script — its responses carry no headers, so
	// no 302 with a Location.
	if meta.Status != http.StatusFound {
		return &SMSMediaDownload{
			Data:          body,
			ContentType:   mediaContentType(meta.Header.Get("Content-Type")),
			ContentLength: int64(len(body)),
		}, nil
	}

	location := meta.Header.Get("Location")
	if location == "" {
		return nil, &apierror.ConnectionError{Err: errors.New("media redirect carried no Location header")}
	}
	data, contentType, err := fetchStored(ctx, timeout, location, "Media")
	if err != nil {
		return nil, err
	}
	return &SMSMediaDownload{Data: data, ContentType: contentType, ContentLength: int64(len(data))}, nil
}
