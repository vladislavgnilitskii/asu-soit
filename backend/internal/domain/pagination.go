package domain

// PageParams — параметры постраничной выборки списков.
type PageParams struct {
	Limit  int
	Offset int
}

// Page — обёртка ответа: страница элементов + общее число записей.
// Total считается с учётом RLS (сколько строк реально видит роль запроса).
type Page[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// NewPage собирает ответ-страницу; nil-срез отдаём как пустой список,
// чтобы клиент всегда получал JSON-массив, а не null.
func NewPage[T any](items []T, total int, p PageParams) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Total: total, Limit: p.Limit, Offset: p.Offset}
}
