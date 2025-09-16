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

	file, expr, err := fileselector.GetRequiredFiles("", parseDate.Format("2006-01-02"), fileArg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type AccountBalance struct {
		Id            string  `json:"id"`
		DisplayName   string  `json:"displayName"`
		Balance       string  `json:"balance"`
		Account       string  `json:"account"`
		PercentChange float64 `json:"percentChange"`
	}

	// Collect all account names
	var accountArgs []string
	for _, sa := range config.Cfg.StarredAccounts {
		accountArgs = append(accountArgs, sa.Account)
	}

	// Helper to run hledger and get balances map
	runHledger := func(endDate string) (map[string]string, error) {
		cmdArgs := []string{"bal", "--no-total", "--end", endDate}
		cmdArgs = append(cmdArgs, accountArgs...)
		for _, f := range file {
			cmdArgs = append(cmdArgs, "-f", f)
		}

		if expr != "" {
			cmdArgs = append(cmdArgs, expr)
		}

		cmd := exec.Command("hledger", cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return nil, fmt.Errorf("hledger error: %w", err)
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		result := make(map[string]string)
		for _, line := range lines {
			for _, acc := range accountArgs {
				if strings.Contains(line, acc) {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						result[acc] = fields[0] + " " + fields[1]
					}
				}
			}
		}
		return result, nil
	}

	// Current month balances
	currentBalances, err := runHledger(date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Previous month balances
	lastMonth := parseDate.AddDate(0, -1, 0).Format("2006-01-02")
	previousBalances, err := runHledger(lastMonth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	var balances []AccountBalance
	for id, sa := range config.Cfg.StarredAccounts {
		cur := currentBalances[sa.Account]
		if cur == "" {
			cur = "0"
		}
		prev := previousBalances[sa.Account]
		if prev == "" {
			prev = "0"
		}

		// Parse numbers (assuming amount currency format like "123.45 USD")
		curFields := strings.Fields(cur)
		prevFields := strings.Fields(prev)

		var curVal, prevVal float64
		if len(curFields) > 0 {
			curVal, _ = strconv.ParseFloat(curFields[0], 64)
		}
		if len(prevFields) > 0 {
			prevVal, _ = strconv.ParseFloat(prevFields[0], 64)
		}

		var pct float64
		if prevVal != 0 {
			pct = ((curVal - prevVal) / prevVal) * 100
		} else {
			pct = 0
		}

		balances = append(balances, AccountBalance{
			Id:            strconv.Itoa(id),
			DisplayName:   sa.DisplayName,
			Balance:       cur,
			Account:       sa.Account,
			PercentChange: pct,
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
