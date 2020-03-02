package csvt

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/taskie/jc"
)

var (
	Version = "0.1.0-beta"
)

type Application struct {
	FromType      string
	ToType        string
	FromDelimiter rune
	ToDelimiter   rune
	Mode          string
	RowRanges     string
	ColRanges     string
}

func NewApplication(mode string) Application {
	return Application{
		Mode:          mode,
		FromType:      "csv",
		ToType:        "csv",
		FromDelimiter: ',',
		ToDelimiter:   ',',
	}
}

func (app *Application) readRecords(r io.Reader) ([][]string, error) {
	switch app.FromType {
	case "csv":
		csvr := csv.NewReader(r)
		csvr.Comma = app.FromDelimiter
		records, err := csvr.ReadAll()
		return records, err
	default:
		var data [][]string
		jcr := jc.NewDecoder(r, app.FromType)
		err := jcr.Decode(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func (app *Application) writeRecords(w io.Writer, records [][]string) error {
	switch app.ToType {
	case "csv":
		csvw := csv.NewWriter(w)
		csvw.Comma = app.ToDelimiter
		return csvw.WriteAll(records)
	default:
		jcw := jc.NewEncoder(w, app.ToType)
		return jcw.Encode(w)
	}
}

func (app *Application) readItems(r io.Reader) ([]map[string]string, error) {
	var data []map[string]string
	jcr := jc.NewDecoder(r, app.FromType)
	err := jcr.Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (app *Application) writeItems(w io.Writer, items []map[string]string) error {
	jcw := jc.NewEncoder(w, app.ToType)
	return jcw.Encode(items)
}

func (app *Application) Convert(r io.Reader, w io.Writer) error {
	return app.convertImpl(r, w, false)
}

func (app *Application) Transpose(r io.Reader, w io.Writer) error {
	return app.convertImpl(r, w, true)
}

func (app *Application) convertImpl(r io.Reader, w io.Writer, transpose bool) error {
	records, err := app.readRecords(r)
	if err != nil {
		return err
	}
	if transpose {
		transposer := Transposer{}
		records, err = transposer.Transpose(records)
		if err != nil {
			return err
		}
	}
	return app.writeRecords(w, records)
}

func (app *Application) Map(r io.Reader, w io.Writer) error {
	records, err := app.readRecords(r)
	if err != nil {
		return err
	}
	mapper := Mapper{}
	items, err := mapper.MapAll(records)
	if err != nil {
		return err
	}
	return app.writeItems(w, items)
}

func (app *Application) Unmap(r io.Reader, w io.Writer) error {
	items, err := app.readItems(r)
	if err != nil {
		return err
	}
	unmapper := Unmapper{}
	records, err := unmapper.UnmapAll(items)
	if err != nil {
		return err
	}
	return app.writeRecords(w, records)
}

func (app *Application) Slice(r io.Reader, w io.Writer) error {
	records, err := app.readRecords(r)
	if err != nil {
		return err
	}
	rowRanges, err := ParseRanges(app.RowRanges)
	if err != nil {
		return err
	}
	colRanges, err := ParseRanges(app.ColRanges)
	if err != nil {
		return err
	}
	slicer := Slicer{
		RowRanges: rowRanges,
		ColRanges: colRanges,
	}
	records, err = slicer.Slice(records)
	if err != nil {
		return err
	}
	return app.writeRecords(w, records)
}

func (app *Application) Run(r io.Reader, w io.Writer) error {
	switch app.Mode {
	case "convert":
		return app.Convert(r, w)
	case "transpose":
		return app.Transpose(r, w)
	case "map":
		return app.Map(r, w)
	case "unmap":
		return app.Unmap(r, w)
	case "slice":
		return app.Slice(r, w)
	default:
		return fmt.Errorf("invalid mode: %s", app.Mode)
	}
}
