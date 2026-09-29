package bird

import (
	"context"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

type AmbConversationsListParams struct {
	BusinessAccountID string                `json:"business_account_id"`
	Status            AMBConversationStatus `json:"status"`
	Queue             *string               `json:"queue"`
	AssignedTo        string                `json:"assigned_to"`
	Label             string                `json:"label"`
	Limit             int                   `json:"limit"`
	EndingBefore      string                `json:"ending_before"`
}

func (p AmbConversationsListParams) toWire(startingAfter string) *oapi.ListAMBConversationsParams {
	return &oapi.ListAMBConversationsParams{
		BusinessAccountId: optZero(p.BusinessAccountID),
		Status:            optZero(p.Status),
		Queue:             p.Queue,
		AssignedTo:        optStr(p.AssignedTo),
		Label:             optStr(p.Label),
		Limit:             optInt(p.Limit),
		EndingBefore:      optStr(p.EndingBefore),
		StartingAfter:     optStr(startingAfter),
	}
}

func (s *AmbConversationsService) ListPage(ctx context.Context, params AmbConversationsListParams, startingAfter string, opts ...option.RequestOption) (*AMBConversationList, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.ListAMBConversations(ctx, params.toWire(startingAfter), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out AMBConversationList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AmbConversationsService) List(ctx context.Context, params AmbConversationsListParams, opts ...option.RequestOption) iter.Seq2[*AMBConversation, error] {
	return paginate(func(cursor string) ([]AMBConversation, *string, error) {
		pageParams := params
		if cursor != "" {
			pageParams.EndingBefore = ""
		}
		page, err := s.ListPage(ctx, pageParams, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
