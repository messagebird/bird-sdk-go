package bird_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"testing"

	bird "github.com/messagebird/bird-sdk-go"
	"github.com/messagebird/bird-sdk-go/option"
)

var attachmentPDF = []byte("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n1 0 obj\n<<>>\nendobj\n")

const attachmentJSON = `{"id":"tca_01krdgeqcxet5s7t44vh8rt9mg","filename":"proof.pdf","content_type":"application/pdf","size_bytes":42,"download_url":"https://files.example/proof.pdf","download_url_expires_at":"2026-10-08T00:00:00Z","status":"draft","expires_at":null,"created_at":"2026-10-07T00:00:00Z","updated_at":"2026-10-07T00:00:00Z"}`

type uploadSeen struct {
	method, path, idempotencyKey string
	filename, description        string
	hasDescription               bool
	data, rawBody                []byte
}

func readUpload(t *testing.T, r *http.Request) uploadSeen {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("Content-Type = %q, want multipart/form-data with a boundary", r.Header.Get("Content-Type"))
	}
	seen := uploadSeen{method: r.Method, path: r.URL.Path, idempotencyKey: r.Header.Get("Idempotency-Key"), rawBody: raw}
	mr := multipart.NewReader(bytes.NewReader(raw), params["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("next part: %v", err)
		}
		value, _ := io.ReadAll(part)
		switch part.FormName() {
		case "file":
			seen.filename, seen.data = part.FileName(), value
		case "description":
			seen.description, seen.hasDescription = string(value), true
		default:
			t.Errorf("unexpected part %q", part.FormName())
		}
	}
	return seen
}

func TestComplianceAttachmentsCreateSendsTheFileAsMultipart(t *testing.T) {
	t.Parallel()
	var seen uploadSeen
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		seen = readUpload(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, attachmentJSON)
	})

	att, err := newClient(t, server).ComplianceAttachments.Create(context.Background(), bird.ComplianceAttachmentsCreateParams{
		File:        bytes.NewReader(attachmentPDF),
		Filename:    "pro\r\nof.pdf",
		Description: bird.Ptr("Utility bill"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if seen.method != http.MethodPost || seen.path != "/v1/compliance/attachments" {
		t.Errorf("request = %s %s, want POST /v1/compliance/attachments", seen.method, seen.path)
	}
	if seen.filename != "proof.pdf" || !bytes.Equal(seen.data, attachmentPDF) {
		t.Errorf("file part = %q (%d bytes), want proof.pdf with the uploaded bytes", seen.filename, len(seen.data))
	}
	if !seen.hasDescription || seen.description != "Utility bill" {
		t.Errorf("description part = %q (present %v), want Utility bill", seen.description, seen.hasDescription)
	}
	if seen.idempotencyKey == "" {
		t.Error("an upload is a mutation and must carry an Idempotency-Key")
	}
	if att.Id != "tca_01krdgeqcxet5s7t44vh8rt9mg" {
		t.Errorf("Id = %q, want the decoded attachment", att.Id)
	}
}

func TestComplianceAttachmentsCreateOmitsAnUnsetDescription(t *testing.T) {
	t.Parallel()
	var seen uploadSeen
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		seen = readUpload(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, attachmentJSON)
	})

	if _, err := newClient(t, server).ComplianceAttachments.Create(context.Background(), bird.ComplianceAttachmentsCreateParams{
		File: bytes.NewReader(attachmentPDF), Filename: "proof.pdf",
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if seen.hasDescription {
		t.Errorf("description part sent as %q, want none", seen.description)
	}
}

// The reader is drained once, so a retry must resend the same body under the
// same key; a second read of the caller's reader would send an empty file.
func TestComplianceAttachmentsCreateRetriesWithTheSameBodyAndKey(t *testing.T) {
	t.Parallel()
	var (
		mu       sync.Mutex
		attempts []uploadSeen
	)
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		seen := readUpload(t, r)
		mu.Lock()
		attempts = append(attempts, seen)
		n := len(attempts)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, `{"error":{"type":"service_unavailable_error","code":"E00503","message":"try again"}}`)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, attachmentJSON)
	})

	if _, err := newClient(t, server, option.WithMaxRetries(1)).ComplianceAttachments.Create(context.Background(), bird.ComplianceAttachmentsCreateParams{
		File: bytes.NewReader(attachmentPDF), Filename: "proof.pdf",
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(attempts) != 2 {
		t.Fatalf("attempts = %d, want 2", len(attempts))
	}
	if attempts[0].idempotencyKey == "" || attempts[0].idempotencyKey != attempts[1].idempotencyKey {
		t.Errorf("Idempotency-Key %q then %q, want one key reused", attempts[0].idempotencyKey, attempts[1].idempotencyKey)
	}
	if !bytes.Equal(attempts[0].rawBody, attempts[1].rawBody) || !bytes.Equal(attempts[1].data, attachmentPDF) {
		t.Error("the retry did not resend the identical multipart body")
	}
}

func TestComplianceAttachmentsGetReturnsTheBytesTypeAndFilename(t *testing.T) {
	t.Parallel()
	var path string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/pdf")
		// The server's form: an ASCII fallback plus the RFC 5987 original.
		w.Header().Set("Content-Disposition", `attachment; filename="_rbol.pdf"; filename*=UTF-8''%C3%A1rbol.pdf`)
		_, _ = w.Write(attachmentPDF)
	})

	file, err := newClient(t, server).ComplianceAttachments.Get(context.Background(), "tca_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if path != "/v1/compliance/attachments/tca_01krdgeqcxet5s7t44vh8rt9mg" {
		t.Errorf("path = %q", path)
	}
	if !bytes.Equal(file.Data, attachmentPDF) || file.ContentType != "application/pdf" || file.Filename != "árbol.pdf" {
		t.Errorf("file = %d bytes, %q, %q; want the PDF, application/pdf, árbol.pdf", len(file.Data), file.ContentType, file.Filename)
	}
}

func TestComplianceDocumentTemplatesGetReadsAWordForm(t *testing.T) {
	t.Parallel()
	const docx = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	var path string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", docx)
		w.Header().Set("Content-Disposition", `attachment; filename="ownership form.docx"`)
		_, _ = w.Write([]byte("PK\x03\x04"))
	})

	file, err := newClient(t, server).ComplianceDocumentTemplates.Get(context.Background(), "cmt_01krdgeqcxet5s7t44vh8rt9mg")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if path != "/v1/compliance/document-templates/cmt_01krdgeqcxet5s7t44vh8rt9mg" {
		t.Errorf("path = %q", path)
	}
	if file.ContentType != docx || file.Filename != "ownership form.docx" || string(file.Data) != "PK\x03\x04" {
		t.Errorf("file = %q, %q, %q", file.Data, file.ContentType, file.Filename)
	}
}

func TestComplianceAttachmentsGetMapsAnAPIError(t *testing.T) {
	t.Parallel()
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"type":"not_found_error","code":"E12044","message":"attachment not found"}}`)
	})

	_, err := newClient(t, server).ComplianceAttachments.Get(context.Background(), "tca_01krdgeqcxet5s7t44vh8rt9mg")
	var apiErr *bird.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v (%T), want *bird.APIError", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "E12044" {
		t.Errorf("StatusCode = %d, Code = %q; want 404 E12044", apiErr.StatusCode, apiErr.Code)
	}
}

func TestComplianceAttachmentsCreateRejectsInvalidInputBeforeSending(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"canceled before read", "canceled during read", "canceled between reads", "oversized"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var reader io.Reader
			switch scenario {
			case "oversized":
				remaining := (10 << 20) + 1
				reader = uploadReaderFunc(func(p []byte) (int, error) {
					if remaining == 0 {
						t.Fatal("read beyond the upload limit plus one byte")
					}
					n := min(len(p), remaining)
					clear(p[:n])
					remaining -= n
					return n, nil
				})
			case "canceled before read":
				cancel()
				reader = uploadReaderFunc(func([]byte) (int, error) {
					t.Fatal("read input after cancellation")
					return 0, io.EOF
				})
			case "canceled between reads":
				reader = uploadReaderFunc(func(p []byte) (int, error) {
					if ctx.Err() != nil {
						t.Fatal("read again after cancellation")
					}
					cancel()
					p[0] = 'x'
					return 1, nil
				})
			case "canceled during read":
				reader = uploadReaderFunc(func([]byte) (int, error) {
					cancel()
					return 0, errors.New("reader failed")
				})
			}
			server := newServer(t, func(http.ResponseWriter, *http.Request) {
				t.Error("sent invalid upload")
			})
			_, err := newClient(t, server).ComplianceAttachments.Create(ctx, bird.ComplianceAttachmentsCreateParams{
				File: reader, Filename: "proof.pdf",
			})
			if scenario == "oversized" {
				if err == nil || !strings.Contains(err.Error(), "10 MiB") {
					t.Fatalf("error = %v, want upload size refusal", err)
				}
			} else if err != context.Canceled {
				t.Fatalf("error = %v, want context.Canceled", err)
			}
		})
	}
}

type uploadReaderFunc func([]byte) (int, error)

func (f uploadReaderFunc) Read(p []byte) (int, error) { return f(p) }
