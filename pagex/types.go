package pagex

const PageLimit = 20

// Deprecated: Use PagePrams instead, is removed in v2.0
type ListReq struct {
	PagePrams
}

func WrapPageParams(page *PagePrams) PagePrams {
	if page == nil {
		return UnLimitPage()
	}
	return *page
}

type PagePrams struct {
	Page     int `json:"page,optional,default=1" form:"page,optional,default=1"`
	PageSize int `json:"pageSize,optional,default=10" form:"pageSize,optional,default=10"`
	LastSize int `json:"lastSize,optional,default=0" form:"lastSize,optional,default=0"`
}

func UnLimitPage() PagePrams {
	return NewPage(1, -1)
}

func NewPage(page, pageSize int) PagePrams {

	return PagePrams{
		Page:     page,
		PageSize: pageSize,
		LastSize: 0,
	}
}

func (page *PagePrams) Limit() int {
	if page.LastSize > 0 {
		return page.LastSize
	}
	if page.PageSize == -1 {
		return -1 // -1 means no limit
	}
	if page.PageSize < 1 {
		return PageLimit
	}
	return page.PageSize
}
func (page *PagePrams) Offset() int {
	if page.PageSize == -1 {
		return 0
	}
	if page.Page == 0 {
		page.Page = 1
	}
	if page.PageSize < 1 {
		page.PageSize = PageLimit
	}
	offset := (page.Page - 1) * page.PageSize
	return offset
}

// Deprecated: Use OrderParams instead, is removed in v2.0
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
