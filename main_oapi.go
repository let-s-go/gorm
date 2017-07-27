package gorm

import (
	"errors"
	"fmt"
)

var (
	createTableCallbacks = make(map[string]func())
)

func RegisterCreateTableCallback(tableName string, callback func()) {
	createTableCallbacks[tableName] = callback
}

func (s *DB) GetCount(where interface{}, args ...interface{}) int {
	cnt := 0
	if where != nil {
		if s.Value == nil && (s.search == nil || s.search.tableName == "") {
			s = s.Model(where)
		} else {
			s = s.Where(where, args...)
		}
	}
	if s.Count(&cnt).Error != nil {
		return -1
	}
	return cnt
}

func (s *DB) DoTransaction(do func(dbc *DB) error) (err error) {
	dbc := s.Begin()
	defer func() {
		if e := recover(); e != nil {
			dbc.Rollback()
			err = errors.New(fmt.Sprint(e))
		}
	}()
	if err := do(dbc); err != nil {
		dbc.Rollback()
		return err
	}
	dbc.Commit()
	return nil
}
