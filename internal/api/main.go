package api

import "net/http"

var fileArg, mainFileArg string

func InitAPI(file, mainFile string) {
	fileArg = file
	mainFileArg = mainFile
	http.HandleFunc("/api/incomestatement/", getIncomeStatement)
	http.HandleFunc("/api/balancesheet/", getBalanceSheet)
	http.HandleFunc("/api/accountBalances/", accountBalances)
	http.HandleFunc("/api/networth/", getNetWorth)
	http.HandleFunc("/api/getConfig/", getConfig)
	http.HandleFunc("/api/updateConfig/", updateConfig)
	http.HandleFunc("/api/sankey/", getSankeyData)
	http.HandleFunc("/api/transactions/", getTransactions)
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}
