package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/A-Bashar/Teka-Finance/internal/config"
	"github.com/A-Bashar/Teka-Finance/internal/fileselector"
)

func accountBalances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	parseDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD.", http.StatusBadRequest)
		return
	}

	file, err := fileselector.GetRequiredFiles("", parseDate.Format("2006-01-02"), fileArg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type AccountBalance struct {
		Id          string `json:"id"`
		DisplayName string `json:"displayName"`
		Balance     string `json:"balance"`
		Account     string `json:"account"`
	}

	// Collect all account names
	var accountArgs []string
	for _, sa := range config.Cfg.StarredAccounts {
		accountArgs = append(accountArgs, sa.Account)
	}

	// Build hledger command: one call for all accounts
	cmdArgs := []string{"bal", "--no-total", "--end", date}
	cmdArgs = append(cmdArgs, accountArgs...)
	for _, f := range file {
		cmdArgs = append(cmdArgs, "-f", f)
	}

	cmd := exec.Command("hledger", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		fmt.Println("Error running hledger:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	var balances []AccountBalance
	for id, sa := range config.Cfg.StarredAccounts {
		// Try to find the line for this account
		var bal string = "0"
		for _, line := range lines {
			if strings.Contains(line, sa.Account) {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					bal = fields[0] + " " + fields[1]
				}
				break
			}
		}
		balances = append(balances, AccountBalance{
			Id:          strconv.Itoa(id),
			DisplayName: sa.DisplayName,
			Balance:     bal,
			Account:     sa.Account,
		})
	}

	jsonResponse, err := json.Marshal(balances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
