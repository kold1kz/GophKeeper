// Package payload содержит структуры полезной нагрузки клиентских данных.
//
// Пакет используется для сериализации разных типов данных перед их
// шифрованием и отправкой на сервер.
package payload

// BinaryFile описывает полезную нагрузку бинарного файла.
//
// Структура используется для хранения имени файла, MIME-типа
// и бинарного содержимого.
type BinaryFile struct {
	// FileName содержит имя файла.
	FileName string `json:"file_name"`
	// MimeType содержит MIME-тип файла.
	MimeType string `json:"mime_type"`
	// Data содержит бинарное содержимое файла.
	Data []byte `json:"data"`
}
