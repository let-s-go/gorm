package gorm

import (
	"strings"

	"github.com/jinzhu/inflection"
)

type RefType int

func (r RefType) String() string {
	if r == RefTypeLeft {
		return "LEFT JOIN"
	}
	return "INNER JOIN"
}

const (
	RefTypeLeft = iota
	RefTypeInner
)

type FieldRef struct {
	Type           RefType
	LocalKey       string
	LocalField     string
	RefTable       string
	RefFluralTable string
	RefKey         string
	RefField       string
}

func (field *StructField) checkFieldRef() {
	if str := field.TagSettings["REF"]; str != "" {
		field.FieldRef = getFieldRef(RefTypeInner, str)
		//		fmt.Println(field.FieldRef)
	} else if str := field.TagSettings["LREF"]; str != "" {
		field.FieldRef = getFieldRef(RefTypeLeft, str)
	}
	if field.FieldRef != nil {
		field.FieldRef.LocalField = ToDBName(field.Name)
		field.IsNormal = false
	}
}
func getFieldRef(typ RefType, str string) *FieldRef {
	var localKey, refTable, refKey, refField string
	//localKey
	if strings.Contains(str, "->") {
		strs := strings.Split(str, "->")
		if len(strs) != 2 {
			return nil
		}
		localKey = strs[0]
		str = strs[1]
	}
	//refKey
	if strings.Contains(str, "(") {
		strs := strings.Split(str, "(")
		if len(strs) != 2 || !strings.HasSuffix(strs[1], ")") {
			return nil
		}
		str = strs[0]
		refKey = strs[1][:len(strs[1])-1]
	} else {
		refKey = "ID"
	}
	//refTable, refField
	strs := strings.Split(str, ".")
	if len(strs) != 2 {
		return nil
	}
	refTable = strs[0]
	refField = strs[1]

	if localKey == "" {
		localKey = refTable + "ID"
	}

	ref := &FieldRef{
		Type:     typ,
		LocalKey: ToDBName(localKey),
		RefTable: ToDBName(refTable),
		RefKey:   ToDBName(refKey),
		RefField: ToDBName(refField),
	}
	ref.RefFluralTable = inflection.Plural(ref.RefTable)
	return ref
}

func (s *DB) Refload(columns ...string) *DB {
	return s.clone().search.Refload(columns...).db
}

func (search *search) Refload(columns ...string) *search {
	search.refload = append(search.refload, columns...)
	return search
}
