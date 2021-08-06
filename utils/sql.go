// Copyright 2021 CloudJ Company Limited. All rights reserved.

package utils

import (
	"bytes"
	"fmt"
	"strings"
)

type BatchSQL struct {
	batchSize int
	offset    int
	op        string
	extraOp   string
	table     string
	columns   []string
	valuesPh  string // VALUES(?,?,...) 占位符
	rowValues [][]interface{}
}

// 建议 batchSize 不要超过 1024
func NewBatchSQL(batchSize int, op string, table string, columns ...string) *BatchSQL {
	ph := make([]string, 0, len(columns))
	for _ = range columns {
		ph = append(ph, "?")
	}

	return &BatchSQL{
		batchSize: batchSize,
		offset:    0,
		op:        op,
		table:     table,
		columns:   columns,
		valuesPh:  fmt.Sprintf("(%s)", strings.Join(ph, ",")),
		rowValues: nil,
	}
}

func (b *BatchSQL) SetTable(table string) {
	b.table = table
}

func (b *BatchSQL) Columns() []string {
	return b.columns
}

func (b *BatchSQL) AddRow(values ...interface{}) error {
	if len(values) != len(b.columns) {
		return fmt.Errorf("got %d row values, expect %d", len(values), len(b.columns))
	}
	b.rowValues = append(b.rowValues, values)
	return nil
}

func (b *BatchSQL) AddExtraOp(op string) {
	if b.extraOp == "" {
		b.extraOp = op
	} else {
		b.extraOp += " " + op
	}
}

func (b *BatchSQL) Reset() {
	b.rowValues = nil
	b.offset = 0
}

func (b *BatchSQL) RowsNum() int {
	return len(b.rowValues)
}

func (b *BatchSQL) HasNext() bool {
	return b.offset < len(b.rowValues)
}

func (b *BatchSQL) Next() (sql string, args []interface{}) {
	start, end := b.offset, b.offset+b.batchSize
	total := len(b.rowValues)
	if start >= total {
		return
	}
	if end > total {
		end = total
	}

	buf := bytes.NewBuffer(nil)
	bPrintf := func(format string, args ...interface{}) {
		_, _ = fmt.Fprintf(buf, format, args...)
	}

	columns := make([]string, len(b.columns))
	for i := range b.columns {
		columns[i] = fmt.Sprintf("`%s`", b.columns[i])
	}

	bPrintf("%s `%s`(%s) VALUES", b.op, b.table, strings.Join(columns, ","))
	for i := range b.rowValues[start:end] {
		if i == 0 {
			bPrintf("%s", b.valuesPh)
		} else {
			bPrintf(",\n%s", b.valuesPh)
		}
	}
	if b.extraOp != "" {
		buf.WriteByte('\n')
		buf.WriteString(b.extraOp)
	}
	buf.WriteByte(';')

	args = make([]interface{}, 0, len(b.rowValues))
	for _, vs := range b.rowValues[start:end] {
		args = append(args, vs...)
	}

	b.offset += b.batchSize
	return buf.String(), args
}
