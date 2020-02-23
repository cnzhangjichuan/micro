package xutils

import (
	"github.com/json-iterator/go"
	"github.com/xlsx"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func MarshalJson(v interface{}) ([]byte, error) {
	switch d := v.(type) {
	default:
		return json.Marshal(v)
	case []byte:
		return d, nil
	}
}

func UnmarshalJson(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func LoadFromFile(conf interface{}, fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, conf)
}

func SplitStrings(s, sep string) []string {
	if s == "" {
		return nil
	}
	if sep == "" {
		sep = ","
	}
	return strings.Split(s, sep)
}

func ParseByte(s string, def byte) byte {
	i, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return def
	}
	return byte(i)
}

func ParseBytes(s string) (ret []byte) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]byte, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseByte(ss[i], 0)
	}
	return
}

func ParseI32(s string, def int32) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return def
	}
	return int32(i)
}

func ParseI32S(s string) (ret []int32) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]int32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseI32(ss[i], 0)
	}
	return
}

func ParseI64(s string, def int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return int64(i)
}

func ParseI64S(s string) (ret []int64) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]int64, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseI64(ss[i], 0)
	}
	return
}

func ParseU32(s string, def uint32) uint32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return def
	}
	return uint32(i)
}

func ParseU32S(s string) (ret []uint32) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]uint32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseU32(ss[i], 0)
	}
	return
}

func ParseU64(s string, def uint64) uint64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return uint64(i)
}

func ParseU64S(s string) (ret []uint64) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]uint64, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseU64(ss[i], 0)
	}
	return
}

func ParseF32(s string, def float32) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return def
	}
	return float32(f)
}

func ParseF32S(s string) (ret []float32) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]float32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseF32(ss[i], 0)
	}
	return
}

func ParseF64(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return float64(f)
}

func ParseF64S(s string) (ret []float64) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]float64, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseF64(ss[i], 0)
	}
	return
}

func ParseBool(s string) bool {
	return s == "1"
}

func ParseBools(s string) (ret []bool) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]bool, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ss[i] == "1"
	}
	return
}

func HasString(arr []string, item string) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func HasU32(arr []uint32, item uint32) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func HasU64(arr []uint64, item uint64) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func ShitLeftStrings(arr []string) []string {
	al := len(arr)
	for i := 1; i < al; i++ {
		arr[i-1] = arr[i]
	}
	return arr[:al-1]
}

func ShitLeftU32S(arr []uint32) []uint32 {
	al := len(arr)
	for i := 1; i < al; i++ {
		arr[i-1] = arr[i]
	}
	return arr[:al-1]
}

type ExcelHelper struct {
	Path   string
	Writer io.Writer
}

// create excel template
func (e *ExcelHelper) CreateTemplate(mode interface{}) error {
	return e.createDataFile(reflect.TypeOf(mode), nil)
}

// write data
func (e *ExcelHelper) Write(data interface{}) error {
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Slice {
		return NewError("must slice param.")
	}

	return e.createDataFile(reflect.TypeOf(data).Elem(), func(fNames []string, sheet *xlsx.Sheet) {
		s := xlsx.NewStyle()
		// font
		s.Font.Name = xlsx.Helvetica
		s.Font.Size = 10
		s.Font.Bold = false

		v := reflect.ValueOf(data)
		for i := 0; i < v.Len(); i++ {
			row := sheet.AddRow()
			e.writeRow(v.Index(i), fNames, row, s)
		}
	})
}

// read data
func (e *ExcelHelper) Read(out interface{}) error {
	xf, err := xlsx.OpenFile(e.Path)
	if err != nil {
		return err
	}
	sheet := xf.Sheets[0]

	t := reflect.TypeOf(out)
	if t.Kind() != reflect.Ptr {
		return NewError("ptr be need.")
	}
	t = t.Elem()
	v := reflect.ValueOf(out)
	ve := v.Elem()

	switch t.Kind() {
	default:
		return NewError("only slice supported!")
	case reflect.Slice:
		et := t.Elem()
		switch et.Kind() {
		default:
			head := e.saxHeadFields(et, sheet.Rows[0].Cells)
			for r := 1; r < len(sheet.Rows); r++ {
				n := reflect.New(et)
				e.readRow(n.Elem(), head, sheet.Rows[r].Cells)
				ve = reflect.Append(ve, n.Elem())
			}
		case reflect.Ptr:
			ett := et.Elem()
			head := e.saxHeadFields(ett, sheet.Rows[0].Cells)
			for r := 1; r < len(sheet.Rows); r++ {
				n := reflect.New(ett)
				e.readRow(n.Elem(), head, sheet.Rows[r].Cells)
				ve = reflect.Append(ve, n)
			}
		}
		v.Elem().Set(ve)
	}
	return nil
}

// wirte data to file
func (e *ExcelHelper) createDataFile(modeType reflect.Type, crDataFunc func([]string, *xlsx.Sheet)) error {
	fd := xlsx.NewFile()

	if modeType.Kind() == reflect.Ptr {
		modeType = modeType.Elem()
	}

	sheet, err := fd.AddSheet("data(" + modeType.Name() + ")")
	if err != nil {
		return err
	}
	head := sheet.AddRow()

	s := xlsx.NewStyle()
	// font
	s.Font.Name = xlsx.Helvetica
	s.Font.Size = 10
	s.Font.Bold = true

	var fNames []string
	for i := 0; i < modeType.NumField(); i++ {
		f := modeType.Field(i)
		tex := f.Tag.Get("xlsx")
		if tex == "" {
			continue
		}
		c := head.AddCell()
		c.SetStyle(s)
		c.SetString(tex)
		fNames = append(fNames, f.Name)
	}

	if crDataFunc != nil {
		crDataFunc(fNames, sheet)
	}

	// writer
	if e.Writer != nil {
		return fd.Write(e.Writer)
	}

	// file
	out, err := os.Create(e.Path)
	if err != nil {
		return err
	}
	err = fd.Write(out)
	out.Close()

	return err
}

func (e *ExcelHelper) saxHeadFields(et reflect.Type, headCells []*xlsx.Cell) []string {
	head := make([]string, len(headCells))
	for c := 0; c < len(headCells); c++ {
		n := headCells[c].String()
		for i := 0; i < et.NumField(); i++ {
			f := et.Field(i)
			if f.Tag.Get("xlsx") == n {
				head[c] = f.Name
				break
			}
		}
	}
	return head
}

func (e *ExcelHelper) writeRow(v reflect.Value, fNames []string, row *xlsx.Row, s *xlsx.Style) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for x := 0; x < len(fNames); x++ {
		f := v.FieldByName(fNames[x])
		if !f.IsValid() {
			continue
		}
		c := row.AddCell()
		c.SetStyle(s)
		o := f.Interface()
		if tm, ok := o.(time.Time); ok {
			c.SetString(tm.Format("2006/01/02 15:04"))
		} else {
			c.SetValue(o)
		}
	}
}

func (e *ExcelHelper) readRow(v reflect.Value, head []string, cells []*xlsx.Cell) {
	for x := 0; x < len(head); x++ {
		if head[x] == "" {
			continue
		}
		f := v.FieldByName(head[x])
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		switch f.Kind() {
		default:
			if f.Type().Name() == "Time" {
				tm, err := cells[x].GetTime(false)
				if err != nil {
					tm = time.Now()
				}
				f.Set(reflect.ValueOf(tm))
			}
		case reflect.String:
			f.SetString(cells[x].Value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if i64, err := cells[x].Int64(); err == nil {
				f.SetInt(i64)
			} else {
				f.SetInt(0)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if u64, err := strconv.ParseUint(cells[x].Value, 10, 64); err == nil {
				f.SetUint(u64)
			} else {
				f.SetUint(0)
			}
		case reflect.Float32, reflect.Float64:
			if f64, err := cells[x].Float(); err == nil {
				f.SetFloat(f64)
			} else {
				f.SetFloat(0)
			}
		case reflect.Bool:
			f.SetBool(cells[x].Bool())
		}
	}
}
