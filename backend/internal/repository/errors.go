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

	ErrInvoiceNotFound   = errors.New("счет не найден")
	ErrInvoiceExists     = errors.New("по заявке счет уже выставлен")
	ErrRequestNotClosed  = errors.New("нельзя выставить счет по незакрытой заявке")
	ErrInvoiceNotPending = errors.New("нельзя менять статус счета, который уже не в статусе pending")

	// ErrEmployeeNotFound — сотрудник с указанным id не существует.
	ErrEmployeeNotFound = errors.New("сотрудник не найден")
	// ErrLoginTaken — логин уже занят (нарушение UNIQUE на employees.login).
	ErrLoginTaken = errors.New("логин уже занят")
	// ErrRoleNotFound — передан несуществующий role_id.
	ErrRoleNotFound = errors.New("роль не найдена")
)
