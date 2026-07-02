package repository

// nullifyEmpty превращает пустую строку в SQL NULL (nil-указатель).
// Нужен для nullable-колонок: пустой ввод не должен попадать в БД как ”,
// иначе теряется разница между «нет значения» и «пустое значение».
func nullifyEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
