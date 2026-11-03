package bird

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

type AssetsUploadsService struct{ resource }

type AssetsUploadsCreateParams struct {
	ClientReference string          `json:"client_reference"`
	File            AssetUploadFile `json:"file"`
	Name            *string         `json:"name,omitempty"`
	AltText         *string         `json:"alt_text,omitempty"`
	FolderID        *string         `json:"folder_id,omitempty"`
}

// Create starts an upload intent for bytes transferred using the returned instruction.
func (s *AssetsUploadsService) Create(ctx context.Context, params AssetsUploadsCreateParams, opts ...option.RequestOption) (*AssetUploadSessionInstruction, error) {
	reference, err := uuid.Parse(params.ClientReference)
	if err != nil {
		return nil, fmt.Errorf("client_reference must be a UUID: %w", err)
	}
	wire := oapi.AssetUploadSessionCreate{ClientReference: reference, File: params.File, Name: params.Name, AltText: params.AltText, FolderId: params.FolderID}
	body, err := s.post(ctx, opts, func(ctx context.Context, idempotencyKey string, cfg requestConfig) (*http.Response, error) {
		op := &oapi.CreateAssetUploadSessionParams{}
		if idempotencyKey != "" {
			op.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateAssetUploadSession(ctx, op, wire, cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AssetUploadSessionInstruction
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
