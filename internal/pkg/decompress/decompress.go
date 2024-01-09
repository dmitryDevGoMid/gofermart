package decompress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
)

// Decompress распаковывает слайс байт.
func DecompressFlate(data []byte) ([]byte, error) {
	// переменная r будет читать входящие данные и распаковывать их
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	var b bytes.Buffer
	// в переменную b записываются распакованные данные
	_, err := b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}

// Распаковываем данные и не забываем закрыть reader gzip
func DecompressGzip(obj []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(obj))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	res, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return res, nil
}
