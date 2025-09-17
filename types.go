package client

type Headers map[string]string

type Params map[string]string

type Href string

type LinksResponse struct {
	Self struct {
		Href `json:"href"`
	} `json:"self"`
	First struct {
		Href `json:"href"`
	} `json:"first"`
	Last struct {
		Href `json:"href"`
	} `json:"last"`
	Prev struct {
		Href `json:"href"`
	} `json:"prev"`
	Next struct {
		Href `json:"href"`
	} `json:"next"`
}

type MetaResponse struct {
	TotalCount  int `json:"totalCount"`
	PageCount   int `json:"pageCount"`
	CurrentPage int `json:"currentPage"`
	PerPage     int `json:"perPage"`
}
