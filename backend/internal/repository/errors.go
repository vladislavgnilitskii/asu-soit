package repository

import "errors"

// Доменные ошибки репозитория, которые хендлеры мапят в конкретные
// HTTP-статусы через errors.Is (вместо «слепого» 500).
var (
	// ErrRequestNotFound — заявка с указанным id не существует.
	ErrRequestNotFound = errors.New("заявка не найдена")
	// ErrStatusNotFound — передан несуществующий status_id.
	ErrStatusNotFound = errors.New("статус не найден")
)
