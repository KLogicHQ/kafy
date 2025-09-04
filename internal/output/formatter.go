package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v3"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

type Formatter struct {
	Format Format
}

func NewFormatter(format string) *Formatter {
	switch format {
	case "json":
		return &Formatter{Format: FormatJSON}
	case "yaml":
		return &Formatter{Format: FormatYAML}
	default:
		return &Formatter{Format: FormatTable}
	}
}

func (f *Formatter) Output(data interface{}) error {
	switch f.Format {
	case FormatJSON:
		return f.outputJSON(data)
	case FormatYAML:
		return f.outputYAML(data)
	default:
		return f.outputTable(data)
	}
}

func (f *Formatter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

func (f *Formatter) outputTable(data interface{}) error {
	// This is a basic implementation - we'll enhance it per command
	fmt.Printf("%+v\n", data)
	return nil
}

func (f *Formatter) OutputTable(headers []string, rows [][]string) {
	if f.Format != FormatTable {
		// For non-table formats, convert to structured data
		result := make([]map[string]string, len(rows))
		for i, row := range rows {
			item := make(map[string]string)
			for j, header := range headers {
				if j < len(row) {
					item[header] = row[j]
				}
			}
			result[i] = item
		}
		f.Output(result)
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	
	// Add headers
	headerRow := make(table.Row, len(headers))
	for i, header := range headers {
		headerRow[i] = header
	}
	t.AppendHeader(headerRow)

	// Add rows
	for _, row := range rows {
		tableRow := make(table.Row, len(row))
		for i, cell := range row {
			tableRow[i] = cell
		}
		t.AppendRow(tableRow)
	}

	t.Render()
}