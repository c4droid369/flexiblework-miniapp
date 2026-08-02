// Package e2e — multipart helpers kept in a separate file so the API
// surface in api_test.go stays focused on assertions.
package e2e

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
)

// multipartWriter builds a minimal multipart/form-data body with one
// "file" part. The caller is responsible for calling Close() before
// reading ContentType.
func multipartWriter(buf *bytes.Buffer, field, filename, content string) *multipart.Writer {
	w := multipart.NewWriter(buf)
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{
		fmt.Sprintf(`form-data; name=%q; filename=%q`, field, filename),
	}
	h["Content-Type"] = []string{"text/plain"}
	p, err := w.CreatePart(h)
	if err != nil {
		panic(err) // test helper — never expected
	}
	if _, err := io.WriteString(p, content); err != nil {
		panic(err)
	}
	return w
}
