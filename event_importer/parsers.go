package main

import (
	"fmt"
	"strconv"
	"reflect"
	"regexp"
	"time"
)


/**
 *	A collection of methods to support marshalling data into some required formats  
 */


func EnsureString(value interface{}) (string) {
	if str, ok := value.(string); ok {
	    return str
	}

    switch value.(type) {	    
	    case int:
	        fmt.Println("Value is int..")
	    case float64:
	        return strconv.Itoa(int(value.(float64)))
	    default:
	        fmt.Println("Conversion of ", reflect.TypeOf(value), " to `string` not implemented.")
    }
    return ""
}

func PluckString(r map[string]interface{}, key string) (string) {
	// I had a couple of issue with json.Number `json:"user_id,Number"` 
	// and, at the time, this seemed more flexible. 
	// I kind of regret not finding something cleaner than these `Pluck` methods.
    return EnsureString(r[key])
}


type CustomTimeParseFormat struct {
	Exp *regexp.Regexp
	Format string
}

// Precompile these
var SomeOperator = CustomTimeParseFormat{
	Exp: regexp.MustCompile(`\d{4}-(0[1-9]|1[0-2])-([0-2][0-9]|3[0-1]{):([0-1][0-9]|2[0-4]):([0-5][0-9]):([0-5][0-9])Z`),
	Format: "2006-01-02:15:04:05Z"} // e.g. 2017-08-20:15:00:00Z  

var AnotherOperator = CustomTimeParseFormat{
	Exp: regexp.MustCompile(`\d{4}-(0[1-9]|1[0-2])-([0-2][0-9]|3[0-1]{) ([0-1][0-9]|2[0-4]):([0-5][0-9]):([0-5][0-9])Z`),
	Format: "2006-01-02 15:04:05Z"} // e.g. 2017-08-28 15:00:00Z

// Unix timestamps are outside the scope of time.Parse 
var rUnix = regexp.MustCompile(`\d{10}`)


func EnsureArray(value interface{}) ([]interface{}) {

	if arr, ok := value.([]interface{}); ok {
		return arr
	}

	return []interface{}{value}

}


func EnsureRFC3339(value interface{}) (string) {
	// Attempts to convert a generic type to an RFC 3339 complient timestamp.
	// Thought I had a nice solution here - just use something external.. 
	// UK mm/dd formats proved problematic though. I didn't want to have to do this switch.

	if strValue, ok := value.(string); ok {
	    switch {
	    	// SomeOperator
		    case SomeOperator.Exp.MatchString(strValue):
		        t, _ := time.Parse(SomeOperator.Format, strValue)
		        return t.Format(time.RFC3339)
		    // AnotherOperator
		    case AnotherOperator.Exp.MatchString(strValue):
		        t, _ :=  time.Parse(AnotherOperator.Format, strValue)
		        return t.Format(time.RFC3339)
		    // Unix
		    case rUnix.MatchString(strValue):
		        i, err := strconv.ParseInt(strValue, 10, 64)
			    if err != nil {panic(err)}
			    return time.Unix(i, 0).Format(time.RFC3339)
			default:
				panic("Can't parse timestamp: " + strValue)
	    }

		return ""

	} 	

    switch value.(type) {
	    case int:
	        return "implement some int->datetime strategy"
	    default:
	        fmt.Println("Conversion of ", reflect.TypeOf(value), " to `date-time` not implemented.")
    }

    return ""

}

func PluckRFC3339(r map[string]interface{}, key string) (string) {
	return EnsureRFC3339(r[key])
}
