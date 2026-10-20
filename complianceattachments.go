package bird

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// ComplianceAttachmentsService uploads, lists, downloads and deletes the
// documents uploaded for registrations. Reach it via Client.ComplianceAttachments.
type ComplianceAttachmentsService struct{ resource }

// ComplianceAttachmentsCreateParams is one document to upload.
type ComplianceAttachmentsCreateParams struct {
	// File is buffered up to the 10 MiB limit so retries resend the same bytes.
	// Cancellation is checked between reads; the reader must return for it to take effect.
	File        io.Reader
	Filename    string
	Description *string
}

// ComplianceFile is a file downloaded from the compliance API: an uploaded
// attachment or a blank form. ContentType and Filename are what the server
// declared for it.
type ComplianceFile struct {
	Data        []byte
	ContentType string
	Filename    string
}

// Create uploads a supporting document for a registration and returns the
// attachment to cite by id in a submission's answers. PNG, JPEG, and PDF files
// up to 10 MB are accepted; the contents determine the type.
//
// Uploading bytes identical to an existing attachment returns that attachment
// instead of a second one, so retrying an upload whose outcome is unknown is
// safe. An attachment no registration cites expires after seven days.
func (s *ComplianceAttachmentsService) Create(ctx context.Context, params ComplianceAttachmentsCreateParams, opts ...option.RequestOption) (*ComplianceAttachment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if params.File == nil {
		return nil, errors.New("bird: ComplianceAttachmentsCreateParams.File is required")
	}
	const maxFileBytes = 10 << 20
	data, err := io.ReadAll(io.LimitReader(complianceUploadReader{ctx: ctx, reader: params.File}, maxFileBytes+1))
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	if err != nil {
		return nil, err
	}
	if len(data) > maxFileBytes {
		return nil, errors.New("bird: compliance attachment exceeds 10 MiB")
	}
	contentType, body, err := attachmentUploadBody(params.Filename, data, params.Description)
	if err != nil {
		return nil, err
	}
	resp, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateComplianceAttachmentParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateComplianceAttachmentWithBody(ctx, op, contentType, bytes.NewReader(body), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out ComplianceAttachment
	if err := decodeBody(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get downloads the bytes of an uploaded registration attachment, with the
// content type and filename it was stored under.
func (s *ComplianceAttachmentsService) Get(ctx context.Context, attachmentID string, opts ...option.RequestOption) (*ComplianceFile, error) {
	body, meta, err := s.getFile(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetComplianceAttachment(ctx, oapi.AttachmentID(attachmentID), nil, cfg...)
	})
	if err != nil {
		return nil, err
	}
	return complianceFile(body, meta.Header), nil
}

// attachmentUploadBody encodes the multipart/form-data body once, so every
// attempt sends the same bytes and boundary, and returns the content type that
// carries the boundary.
func attachmentUploadBody(filename string, data []byte, description *string) (string, []byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", strings.NewReplacer("\r", "", "\n", "").Replace(filename))
	if err != nil {
		return "", nil, err
	}
	if _, err := part.Write(data); err != nil {
		return "", nil, err
	}
	if description != nil {
		if err := w.WriteField("description", *description); err != nil {
			return "", nil, err
		}
	}
	if err := w.Close(); err != nil {
		return "", nil, err
	}
	return w.FormDataContentType(), buf.Bytes(), nil
}

func complianceFile(data []byte, header http.Header) *ComplianceFile {
	file := &ComplianceFile{Data: data, ContentType: mediaContentType(header.Get("Content-Type"))}
	if _, params, err := mime.ParseMediaType(header.Get("Content-Disposition")); err == nil {
		file.Filename = params["filename"]
	}
	return file
}

type complianceUploadReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r complianceUploadReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}
