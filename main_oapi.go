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

// Find find records that match given conditions
func (s *DB) Page(out interface{}, where ...interface{}) (cnt int64, err error) {
	if s.search != nil && fmt.Sprint(s.search.offset) == "0" {
		limit := s.search.limit
		s.search.limit = nil
		s.search.offset = nil
		if err := s.Model(out).Count(&cnt).Error; err != nil {
			return 0, err
		}
		if cnt == 0 {
			return 0, nil
		}
		s.search.limit = limit
		s.search.offset = 0
	}
	return cnt, s.clone().NewScope(out).inlineCondition(where...).callCallbacks(s.parent.callbacks.queries).db.Error
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
