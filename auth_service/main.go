package main

import (
	db_adapter "db_adapter/src"
	"fmt"
)

func main() {
	connect_options := db_adapter.PostgresConnectOptions{
		UserName:     "postgres",
		UserPassword: "1234",
		Host:         "localhost",
		Port:         5432,
		DBName:       "Kuznetsov",
	}
	adapter, err := db_adapter.PostgresConnect(connect_options)

	if err == nil {
		fmt.Printf("Connected Successfully with options: %#v\n", connect_options)
	} else {
		fmt.Println("Error connecting to Database", err)
	}

	_ = adapter

	fmt.Println("Hello from auth service")
	fmt.Println("Hello again")
}
