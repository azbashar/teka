package api

import "net/http"

var fileArg, mainFileArg string

func InitAPI(file, mainFile string) {
	fileArg = file
	mainFileArg = mainFile
	http.HandleFunc("/api/incomestatement/", getIncomeStatement)
}