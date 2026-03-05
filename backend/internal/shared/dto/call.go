package dto

import (
	"telephony/internal/domain"
	"telephony/models"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
)

// CallsListRequestToFilters конвертирует models.CallsListRequest в domain.CallListFilters.
func CallsListRequestToFilters(companyID uuid.UUID, req *models.CallsListRequest) *domain.CallListFilters {
	filters := &domain.CallListFilters{
		CompanyID: companyID,
	}
	if req == nil || req.Page == nil {
		return filters
	}
	if req.Page.Limit != nil {
		filters.Limit = int(*req.Page.Limit)
	}
	if req.Page.Offset != nil {
		filters.Offset = int(*req.Page.Offset)
	}
	if req.Page.Page != nil {
		filters.Page = int(*req.Page.Page)
	}

	if t := time.Time(req.DateFrom); !t.IsZero() {
		from := t
		filters.From = &from
	}
	if t := time.Time(req.DateTo); !t.IsZero() {
		end := t.Add(24 * time.Hour)
		filters.To = &end
	}

	if req.Direction != "" {
		d := domain.CallDirection(req.Direction)
		filters.Direction = &d
	}
	if req.Status != "" {
		s := domain.CallEventStatus(req.Status)
		filters.Status = &s
	}
	if s := req.CompanyTelephonyID.String(); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			filters.CompanyTelephonyID = &id
		}
	}

	return filters
}

func callEventToResponse(ev *domain.CallEvent) *models.CallEventResponse {
	if ev == nil {
		return nil
	}
	return &models.CallEventResponse{
		ID:        strfmt.UUID(ev.ID.String()),
		Status:    string(ev.Status),
		Timestamp: strfmt.DateTime(ev.Timestamp),
	}
}

// CallSummaryDomainToModel конвертирует domain.CallSummary в models.CallsListItem.
func CallSummaryDomainToModel(c *domain.CallSummary) *models.CallsListItem {
	if c == nil {
		return nil
	}
	return &models.CallsListItem{
		ID:                 strfmt.UUID(c.ID.String()),
		CompanyTelephonyID: strfmt.UUID(c.CompanyTelephonyID.String()),
		FromNumber:         c.FromNumber,
		ToNumber:           c.ToNumber,
		Direction:          string(c.Direction),
		CreatedAt:          strfmt.DateTime(c.CreatedAt),
		UpdatedAt:          strfmt.DateTime(c.UpdatedAt),
		LastStatus:         string(c.LastStatus),
		HasChildren:        c.HasChildren,
	}
}

// CallTreeDomainToModel конвертирует domain.CallTree в models.CallTreeResponse.
func CallTreeDomainToModel(tree *domain.CallTree) *models.CallTreeResponse {
	if tree == nil || tree.Call == nil {
		return nil
	}
	return &models.CallTreeResponse{
		Call:     callToCallResponse(tree.Call),
		Children: MapSlice(tree.Children, CallTreeDomainToModel),
	}
}

func callToCallResponse(call *domain.Call) *models.CallResponse {
	if call == nil {
		return nil
	}
	parentID := strfmt.UUID("")
	if call.ParentCallID != uuid.Nil {
		parentID = strfmt.UUID(call.ParentCallID.String())
	}
	events := MapSlice(call.Events, callEventToResponse)
	var details *models.CallDetailsResponse
	if call.Details != nil {
		d := call.Details
		details = &models.CallDetailsResponse{
			RecordingSid:      d.RecordingSid,
			RecordingURL:      d.RecordingURL,
			RecordingDuration: int32(d.RecordingDuration),
			FromCountry:       d.FromCountry,
			FromCity:          d.FromCity,
			ToCountry:         d.ToCountry,
			ToCity:            d.ToCity,
			Carrier:           d.Carrier,
			Trunk:             d.Trunk,
		}
	}
	return &models.CallResponse{
		ID:                   strfmt.UUID(call.ID.String()),
		CompanyTelephonyID:   strfmt.UUID(call.CompanyTelephonyID.String()),
		ParentCallID:         parentID,
		ExternalCallID:       call.ExternalCallID,
		ExternalParentCallID: call.ExternalParentCallID,
		FromNumber:           call.FromNumber,
		ToNumber:             call.ToNumber,
		Direction:            string(call.Direction),
		CreatedAt:            strfmt.DateTime(call.CreatedAt),
		UpdatedAt:            strfmt.DateTime(call.UpdatedAt),
		Events:               events,
		Details:              details,
	}
}
