package pagex

const PageLimit = 20

type ListReq struct {
	PagePrams
}
type PagePrams struct {
	Page     int `json:"page,optional,default=1" form:"page,optional,default=1"`
	PageSize int `json:"pageSize,optional,default=10" form:"pageSize,optional,default=10"`
	LastSize int `json:"lastSize,optional,default=0" form:"lastSize,optional,default=0"`
}

func (page *PagePrams) Limit() int {
	if page.LastSize > 0 {
		return page.LastSize
	}
	if page.PageSize < 1 {
		return PageLimit
	}
	return page.PageSize
}
func (page *PagePrams) Offset() int {
	if page.Page == 0 {
		page.Page = 1
	}
	if page.PageSize < 1 {
		page.PageSize = PageLimit
	}
	offset := (page.Page - 1) * page.PageSize
	return offset
}

type OrderBy OrderParams
type OrderParams struct {
	OrderKey string `json:"orderKey"`
	Sort     string `json:"sort"`
}

func EmptyOrderBy() OrderParams {
	return OrderParams{
		OrderKey: "",
		Sort:     "",
	}
}
