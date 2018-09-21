package csvt

import (
	"encoding/csv"
	"fmt"
	"github.com/taskie/jc"
	"io"
)

var (
	Version  = "0.1.0-beta"
	Revision = ""
)

type Application struct {
	FromType      string
	ToType        string
	FromDelimiter rune
	ToDelimiter   rune
	Mode          string
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
		jcr := jc.Jc{
			FromType: app.FromType,
		}
		err := jcr.Decode(r, &data)
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
		jcw := jc.Jc{
			ToType: app.ToType,
		}
		return jcw.Encode(w, records)
	}
}

func (app *Application) readItems(r io.Reader) ([]map[string]string, error) {
	var data []map[string]string
	jcr := jc.Jc{
		FromType: app.FromType,
	}
	err := jcr.Decode(r, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (app *Application) writeItems(w io.Writer, items []map[string]string) error {
	jcw := jc.Jc{
		ToType: app.ToType,
	}
	return jcw.Encode(w, items)
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
	default:
		return fmt.Errorf("invalid mode: %s", app.Mode)
	}
}
