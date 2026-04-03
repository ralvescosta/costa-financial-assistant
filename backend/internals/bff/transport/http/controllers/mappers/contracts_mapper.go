package mappers

import bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"

const defaultPageSize int32 = 25

// ToServicePage normalizes optional query pagination into a service contract.
func ToServicePage(pageSize int32, pageToken string) bffcontracts.Page {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	return bffcontracts.Page{
		Size:  pageSize,
		Token: pageToken,
	}
}

// ToTransportNextPageToken converts a service page result to a transport token.
func ToTransportNextPageToken(page *bffcontracts.PageResult) string {
	if page == nil {
		return ""
	}

	return page.NextToken
}
