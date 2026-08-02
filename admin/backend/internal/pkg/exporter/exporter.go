// Package exporter writes tabular data to Excel or CSV. Callers supply a
// header row and a slice of rows; the exporter handles streaming the bytes
// back to the client.
package exporter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// Sheet is the in-memory representation of a workbook. One sheet per file;
// multi-sheet exports can call NewSheet multiple times in the future.
type Sheet struct {
	Name    string
	Headers []string
	Rows    [][]any
}

// Excel writes the workbook to the response and ends the request.
func Excel(c *gin.Context, filename string, sheets ...Sheet) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	for i, s := range sheets {
		sheet := s.Name
		if sheet == "" {
			sheet = "Sheet" + strconv.Itoa(i+1)
		}
		if i == 0 {
			_ = f.SetSheetName("Sheet1", sheet)
		} else {
			_, _ = f.NewSheet(sheet)
		}
		writeSheet(f, sheet, s)
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_%s.xlsx"`,
		filename, time.Now().UTC().Format("20060102_150405")))
	return f.Write(c.Writer)
}

func writeSheet(f *excelize.File, sheet string, s Sheet) {
	for col, h := range s.Headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for r, row := range s.Rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
}

// CSV writes one sheet as CSV to the response.
func CSV(c *gin.Context, filename string, s Sheet) error {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(s.Headers); err != nil {
		return err
	}
	for _, row := range s.Rows {
		rec := make([]string, len(row))
		for i, v := range row {
			rec[i] = stringify(v)
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_%s.csv"`,
		filename, time.Now().UTC().Format("20060102_150405")))
	c.Status(http.StatusOK)
	_, err := c.Writer.Write(buf.Bytes())
	return err
}

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case time.Time:
		return x.Format(time.RFC3339)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", x)
	}
}
