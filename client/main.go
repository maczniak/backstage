package main

import (
	. "backstage/postgresql"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf(`"config.yaml" file error: %s\n`, err))
	}
	server := viper.GetString("server")
	user := viper.GetString("user")
	password := viper.GetString("password")
	database := viper.GetString("database")
	query := viper.GetString("query")

	conn, err := Connect(server)
	checkError(err)

	params, err := conn.Login(user, password, database)
	checkError(err)
	if viper.GetBool("debug") {
		fmt.Println(params)
		fmt.Println()
	}

	results, err := conn.Query(query)
	checkError(err)
	if viper.GetBool("debug") {
		fmt.Println(results["description"])
		fmt.Println(StringRows(results["rows"].([]DataRow),
			results["description"].(RowDescription)))
		fmt.Println(results["command_tag"])
		fmt.Println(results["transaction_status"])
		fmt.Println()
	}

	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
