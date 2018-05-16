// Copyright 2016 Fraunhofer Institute for Applied Information Technology FIT

package common

import (
	"regexp"
	"strings"
)

const (
	// IDSeparator is used for separation of IDs in the URL
	IDSeparator = ","

	// Location of APIs
	RegistryAPILoc = "/registry"
	DataAPILoc     = "/data"
	AggrAPILoc     = "/aggr"

	// Query parameters
	ParamPage    = "page"
	ParamPerPage = "per_page"
	ParamLimit   = "limit"
	ParamStart   = "start"
	ParamEnd     = "end"
	ParamSort    = "sort"
	// Values for ParamSort
	ASC  = "asc"  // ascending
	DESC = "desc" // descending
)

// Data source types
const (
	STRING = "string"
	FLOAT  = "float"
	BOOL   = "bool"
)

var (
	// APIVersion defines the API version
	APIVersion = "N/A"
	// Default MIME type for all responses
	DefaultMIMEType = "application/vnd.eu.linksmart.hds+json;version=" + APIVersion

	// supported type values
	supportedTypes = []string{STRING, BOOL, FLOAT}
	// supported aggregates
	supportedAggregates = []string{"mean", "stddev", "sum", "min", "max", "median"}
	// supported period suffixes
	supportedPeriods = []string{"m", "h", "d", "w"}
)

// SupportedPeriod validates a period
func SupportedPeriod(p string) bool {
	if p == "" {
		// empty means no retention
		return true
	}
	// Create regexp: ^[0-9]*(h|d|w|m)$
	intervals := strings.Join(supportedPeriods, "|")
	re := regexp.MustCompile("^[0-9]*(" + intervals + ")$")
	return re.MatchString(p)
}

// SupportedType validates a type
func SupportedType(t string) bool {
	return stringInSlice(t, supportedTypes)
}

// SupportedAggregate validates an aggregate
func SupportedAggregate(a string) bool {
	return stringInSlice(a, supportedAggregates)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
