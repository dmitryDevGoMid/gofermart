package gzipandunserialize

import (
	"fmt"
	"io"

	"github.com/dmitryDevGoMid/gofermart/internal/pkg/decompress"
	"github.com/dmitryDevGoMid/gofermart/internal/pkg/pipeline"
	"github.com/dmitryDevGoMid/gofermart/internal/service"
)

type Gzip struct{}

// func (chain *CheckGzip) run(r *Request) error {
func (m Gzip) Process(result pipeline.Message) ([]pipeline.Message, error) {

	fmt.Println("Processing Gzip")

	data := result.(*service.Data)

	compress_ := false

	content := data.Default.Ctx.Request.Header.Values("Content-Encoding")

	for _, val := range content {
		if val == "gzip" {
			compress_ = true
		}
	}

	body, err := io.ReadAll(data.Default.Ctx.Request.Body)

	if err != nil {
		return []pipeline.Message{data}, err
	}

	if compress_ {

		decompr, _ := decompress.DecompressGzip(body)

		data.Default.Body = decompr

		return []pipeline.Message{data}, nil
	}

	data.Default.Body = body

	fmt.Println("BODY====>", data.Default.Body)

	return []pipeline.Message{data}, nil
}
