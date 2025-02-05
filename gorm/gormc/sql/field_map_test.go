package sql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDbName(t *testing.T) {
	tag := "autoUpdateTime:milli;column:id"
	field, err := getDbField(tag)
	assert.Nil(t, err, "err not nil")
	assert.Equal(t, "id", field)
}

type TestStruct struct {
	F1 string    `gorm:"column:f1"`
	F2 int       `gorm:"column:f2"`
	F3 bool      `gorm:"column:f3"`
	F4 time.Time `gorm:"column:f4"`
}

func (ts TestStruct) TableName() string {
	return "table_test"
}

var d TestStruct

func TestGetDbNameFromStruct(t *testing.T) {
	InitField(&d)

	assert.Equal(t, "table_test.f1", Field(&d.F1))
}
