package gorm

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
