package bird

import (
	"context"
	"errors"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/apierror"
	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// VoiceLegsRecordingsService reaches the recordings of one leg. Reach it via
// Client.Voice.Legs.Recordings.
type VoiceLegsRecordingsService struct{ resource }

// VoiceRecordingAudio is the downloaded audio of one recording. ContentType is
// what storage declared, audio/wav.
type VoiceRecordingAudio struct {
	Data          []byte
	ContentType   string
	ContentLength int64
}

// Audio downloads one recording as a WAV file. legID and recordingID are the
// leg_id and id of an entry in Recordings, which Client.Voice.Calls.Recordings.List
// returns.
//
// A 425 *APIError means the leg is still recording: call again in a few
// seconds. A 404 *APIError means the capture failed, or the recording is older
// than 30 days.
//
// The second leg runs on http.DefaultClient rather than the client's own:
// option.WithHTTPClient may carry auth middleware, and the presigned target
// refuses a second auth mechanism.
func (s *VoiceLegsRecordingsService) Audio(ctx context.Context, legID, recordingID string, opts ...option.RequestOption) (*VoiceRecordingAudio, error) {
	body, meta, err := s.getRedirect(ctx, opts, http.StatusFound, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapiNoRedirect.GetVoiceLegRecordingAudio(ctx, oapi.VoiceCallID(legID), oapi.VoiceCallRecordingID(recordingID), nil, cfg...)
	})
	if err != nil {
		return nil, err
	}
	contentType := mediaContentType(meta.Header.Get("Content-Type"))
	// A 2xx is an edge answering with the bytes directly, which is also the only
	// arm the conformance corpus can script: its responses carry no headers, so
	// no 302 with a Location.
	if meta.Status == http.StatusFound {
		location := meta.Header.Get("Location")
		if location == "" {
			return nil, &apierror.ConnectionError{Err: errors.New("recording redirect carried no Location header")}
		}
		if body, contentType, err = fetchStored(ctx, location, "Audio"); err != nil {
			return nil, err
		}
	}
	return &VoiceRecordingAudio{Data: body, ContentType: contentType, ContentLength: int64(len(body))}, nil
}
