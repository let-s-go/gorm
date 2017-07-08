package gorm

import (
	"fmt"
	"reflect"
	"strings"
)

func refloadCallback(scope *Scope) {
	if scope.Search.selects != nil || scope.Search.joinConditions != nil || scope.Search.refload == nil || scope.HasError() {
		return
	}

	var fieldRefs []*FieldRef
	for _, col := range scope.Search.refload {
		find := false
		for _, field := range scope.Fields() {
			if field.FieldRef != nil && (col == field.Name || col == ".") {
				fieldRefs = append(fieldRefs, field.FieldRef)
				//				scope.handleRefload(field)
				find = true
			}
		}
		if !find {
			scope.Err(fmt.Errorf("can't refload field %s for %s", col, scope.GetModelStruct().ModelType))
		}
	}
	if len(fieldRefs) > 0 {
		joins := make(map[string]string)
		selects := fmt.Sprintf("%v.*", scope.QuotedTableName())
		for _, ref := range fieldRefs {
			refTableName := scope.Quote(getTableName(scope, ref.RefTable, ref.RefFluralTable))
			selects += "," + fmt.Sprintf("%v.%v %v", refTableName, scope.Quote(ref.RefField), scope.Quote(ref.LocalField))

			join := fmt.Sprintf("%v %v ON %v.%v=%v.%v", ref.Type, refTableName,
				scope.QuotedTableName(), scope.Quote(ref.LocalKey), refTableName, scope.Quote(ref.RefKey))

			if str, ok := joins[ref.RefTable]; !ok || strings.HasPrefix(str, "LEFT") {
				joins[ref.RefTable] = join
			}
		}
		scope.Search.Select(selects)
		for _, join := range joins {
			scope.Search.Joins(join)
		}
	}
}

func (scope *Scope) handleRefload(field *Field) {
	ref := field.FieldRef

	primaryKeys := scope.getColumnAsArray([]string{ref.LocalKey}, scope.Value)
	if len(primaryKeys) == 0 {
		return
	}
	tableName := getTableName(scope, ref.RefTable, ref.RefFluralTable)
	selects := ToDBName(ref.RefKey) + "," + ToDBName(ref.RefField)
	query := fmt.Sprintf("%v IN (%v)", toQueryCondition(scope, []string{ToDBName(ref.RefKey)}), toQueryMarks(primaryKeys))
	values := toQueryValues(primaryKeys)

	rows, err := scope.NewDB().Table(tableName).Where(query, values...).Select(selects).Rows()

	if err != nil {
		scope.Err(err)
		return
	}
	results := make(map[string]interface{})
	for rows.Next() {
		var k string
		v := reflect.New(field.Struct.Type).Interface()

		rows.Scan(&k, v)
		results[k] = v
	}
	// assign find results
	var (
		indirectScopeValue = scope.IndirectValue()
	)

	if indirectScopeValue.Kind() == reflect.Slice {
		for i := 0; i < indirectScopeValue.Len(); i++ {
			object := indirect(indirectScopeValue.Index(i))
			f := object.FieldByName(field.Name)
			k := fmt.Sprintf("%v", object.FieldByName(ref.LocalKey).Interface())
			if v, ok := results[k]; ok {
				f.Set(indirect(reflect.ValueOf(v)))
			}
		}
	} else {
		k := fmt.Sprintf("%v", indirectScopeValue.FieldByName(ref.LocalKey).Interface())
		if v, ok := results[k]; ok {
			field.Set(indirect(reflect.ValueOf(v)))
		}
	}
}

func getTableName(scope *Scope, name string, fluralName string) string {
	if scope.db == nil || !scope.db.parent.singularTable {
		return fluralName
	}
	return name
}
