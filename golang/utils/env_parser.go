package utils

import (
	"log"
	"os"
	"strconv"
)

func ReadEnv(variable_name string, default_value string) string {
	val := os.Getenv(variable_name)
	if val == "" {
		log.Printf("Env variable %s is not defined. Default value %s default_value will be used", variable_name, default_value)
		val = default_value
	}
	return val
}

func ReadEnvU16(variable_name string, default_value uint16) uint16 {
	val := os.Getenv(variable_name)
	var res uint16
	if val == "" {
		log.Printf("Env variable %s is not defined. Default value %d default_value will be used", variable_name, default_value)
		res = default_value
	} else {
		tmp, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			log.Printf("Env variable %s is defined but is not a string. Default value %d will be used", variable_name, default_value)
			res = default_value
		} else {
			res = uint16(tmp)
		}
	}
	return res
}
