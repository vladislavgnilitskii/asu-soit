package repository

import "errors"

// Доменные ошибки репозитория, которые хендлеры мапят в конкретные
// HTTP-статусы через errors.Is (вместо «слепого» 500).
var (
	// ErrRequestNotFound — заявка с указанным id не существует.
	ErrRequestNotFound = errors.New("заявка не найдена")
	// ErrStatusNotFound — передан несуществующий status_id.
	ErrStatusNotFound = errors.New("статус не найден")

	// ErrPartNotFound — запчасть с указанным id не существует.
	ErrPartNotFound = errors.New("запчасть не найдена")
	// ErrCategoryNotFound — передан несуществующий category_id.
	ErrCategoryNotFound = errors.New("категория запчастей не найдена")
	// ErrInsufficientStock — на складе меньше деталей, чем запрошено к списанию/выдаче.
	ErrInsufficientStock = errors.New("недостаточно деталей на складе")
	// ErrAlreadyIssued — эта деталь уже списана на данную заявку
	// (нарушение uq_request_parts на стороне БД).
	ErrAlreadyIssued = errors.New("деталь уже списана на эту заявку")
)
