package mappers

import (
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

func ToUploadRequest(input *views.UploadDocumentInput) (string, []byte) {
	if input == nil {
		return "", nil
	}
	return input.FileName, input.RawBody
}

func ToClassifyRequest(input *views.ClassifyDocumentInput) (string, string) {
	if input == nil {
		return "", ""
	}
	return input.DocumentID, input.Body.Kind
}

func ToListDocumentsRequest(input *views.ListDocumentsInput) (int32, string) {
	if input == nil {
		page := ToServicePage(0, "")
		return page.Size, page.Token
	}

	page := ToServicePage(input.PageSize, input.PageToken)
	return page.Size, page.Token
}

func ToGetDocumentRequest(input *views.GetDocumentInput) string {
	if input == nil {
		return ""
	}
	return input.DocumentID
}

func ToDocumentResponse(resp *bffcontracts.DocumentResponse) views.DocumentResponse {
	if resp == nil {
		return views.DocumentResponse{}
	}
	return *resp
}

func ToListDocumentsResponse(resp *bffcontracts.ListDocumentsResponse) views.ListDocumentsResponse {
	if resp == nil {
		return views.ListDocumentsResponse{Items: []*views.DocumentResponse{}}
	}
	return *resp
}

func ToDocumentDetailResponse(resp *bffcontracts.DocumentDetailResponse) views.DocumentDetailResponse {
	if resp == nil {
		return views.DocumentDetailResponse{}
	}
	return *resp
}
