package main

import (
	"fmt"
	"net/url"
	"strings"
)

type SQLModifier struct {
	SortField string
	SortAsc   bool
}

func ValuesToModifier(values url.Values, columns map[string]string) SQLModifier {
	m := SQLModifier{}
	sortArray, ok := values["sort"]
	if !ok {
		log.Debug("No sort key")
		return m
	}
	// use only first sort key
	s := strings.Trim(sortArray[0], "[]\"")
	a := strings.Split(s, "\",\"")
	if len(a) != 2 {
		return m
	}
	switch strings.ToLower(a[1]) {
	case "asc":
		m.SortAsc = true
	case "desc":
		m.SortAsc = false
	default:
		return m
	}
	translated, ok := columns[a[0]]
	if !ok {
		return m
	}
	m.SortField = translated
	return m
}

func BuildQuery(base string, m SQLModifier) string {
	var query, order string

	query = base

	if m.SortField == "" {
		return query
	}
	if m.SortAsc {
		order = "ASC"
	} else {
		order = "DESC"
	}
	query = fmt.Sprintf("%s ORDER BY %s %s", query, m.SortField, order)
	log.Debug("Built query", "query", query, "modifier", m)
	return query
}
