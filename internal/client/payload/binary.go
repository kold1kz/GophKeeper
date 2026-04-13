package payload

type BinaryFile struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	Data     []byte `json:"data"`
}
