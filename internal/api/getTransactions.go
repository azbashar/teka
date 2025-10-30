package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/azbashar/teka/internal/config"
	"github.com/azbashar/teka/internal/fileselector"
)

func getTransactions(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	account := r.URL.Query().Get("account")
	value := r.URL.Query().Get("valueMode")
	cost := r.URL.Query().Get("cost")

	// Build hledger command
	cmdArgs := []string{"print", "-O", "json"}

	if startDate != "" {
		cmdArgs = append(cmdArgs, "-b", startDate)
	}
	if endDate != "" {
		cmdArgs = append(cmdArgs, "-e", endDate)
	}
	if account != "" {
		cmdArgs = append(cmdArgs, account)
	}
	if value == "then" || value == "now" || value == "end" {
		cmdArgs = append(cmdArgs, "--value="+value+","+config.Cfg.BaseCurrency)
	} else if value != "" {
		http.Error(w, "Invalid value mode. Allowed options are then/now/end.", http.StatusBadRequest)
		return
	}
	if cost == "true" {
		cmdArgs = append(cmdArgs, "--cost")
	}

	files, expr, err := fileselector.GetRequiredFiles(startDate, endDate, fileArg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, f := range files {
		cmdArgs = append(cmdArgs, "-f", f)
	}
	if expr != "" {
		cmdArgs = append(cmdArgs, expr)
	}

	// Run hledger
	out, err := exec.Command("hledger", cmdArgs...).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		http.Error(w, err.Error()+": "+string(out), http.StatusInternalServerError)
		return
	}

	// Parse JSON into generic slice
	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse hledger output: %v", err), http.StatusInternalServerError)
		return
	}

	type Tag struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	type Cost struct {
		HasCost   bool    `json:"hasCost"`
		Amount    float64 `json:"amount"`
		Commodity string  `json:"commodity"`
	}
	type Posting struct {
		Account   string  `json:"account"`
		Amount    float64 `json:"amount"`
		Commodity string  `json:"commodity"`
		Cost      Cost    `json:"cost"`
		Comment   string  `json:"comment"`
		Status    string  `json:"status"`
		Tags      []Tag   `json:"tags"`
	}
	type Doc struct {
		Attached bool   `json:"attached"`
		Path     string `json:"path"`
	}
	type Transaction struct {
		ID          int       `json:"id"`
		Date        string    `json:"date"`
		Description string    `json:"description"`
		Tags        []Tag     `json:"tags"`
		Comment     string    `json:"comment"`
		Code        string    `json:"code"`
		Status      string    `json:"status"`
		Doc         Doc       `json:"doc"`
		Postings    []Posting `json:"postings"`
	}
	resp := struct {
		Transactions []Transaction `json:"transactions"`
	}{}

	for _, tx := range raw {
		tags := []Tag{}
		if tt, ok := tx["ttags"].([]any); ok {
			for _, kv := range tt {
				if pair, ok := kv.([]any); ok && len(pair) == 2 {
					key, _ := pair[0].(string)
					val, _ := pair[1].(string)
					tags = append(tags, Tag{Key: key, Value: val})
				}
			}
		}

		doc := Doc{Attached: false}
		for _, t := range tags {
			if t.Key == "doc" {
				doc.Attached = true
				doc.Path = t.Value
			}
		}

		postings := []Posting{}
		if ps, ok := tx["tpostings"].([]any); ok {
			for _, p := range ps {
				if pmap, ok := p.(map[string]any); ok {
					acc, _ := pmap["paccount"].(string)
					status, _ := pmap["pstatus"].(string)
					comment, _ := pmap["pcomment"].(string)

					amount := 0.0
					commodity := ""
					costData := Cost{HasCost: false}

					if amounts, ok := pmap["pamount"].([]any); ok && len(amounts) > 0 {
						if am, ok := amounts[0].(map[string]any); ok {
							if aq, ok := am["aquantity"].(map[string]any); ok {
								amount, _ = aq["floatingPoint"].(float64)
							}
							if comm, ok := am["acommodity"].(string); ok {
								commodity = comm
							}
							if c, ok := am["acost"].(map[string]any); ok {
								if inner, ok := c["contents"].(map[string]any); ok {
									cc := Cost{HasCost: true}
									if aq, ok := inner["aquantity"].(map[string]any); ok {
										cc.Amount, _ = aq["floatingPoint"].(float64)
									}
									if comm, ok := inner["acommodity"].(string); ok {
										cc.Commodity = comm
									}
									costData = cc
								}
							}
						}
					}

					// posting tags
					ptags := []Tag{}
					if tlist, ok := pmap["ptags"].([]any); ok {
						for _, kv := range tlist {
							if pair, ok := kv.([]any); ok && len(pair) == 2 {
								key, _ := pair[0].(string)
								val, _ := pair[1].(string)
								ptags = append(ptags, Tag{Key: key, Value: val})
							}
						}
					}

					postings = append(postings, Posting{
						Account:   acc,
						Amount:    amount,
						Commodity: commodity,
						Cost:      costData,
						Comment:   comment,
						Status:    status,
						Tags:      ptags,
					})
				}
			}
		}

		resp.Transactions = append(resp.Transactions, Transaction{
			ID:          int(tx["tindex"].(float64)),
			Date:        tx["tdate"].(string),
			Description: tx["tdescription"].(string),
			Tags:        tags,
			Comment:     tx["tcomment"].(string),
			Code:        tx["tcode"].(string),
			Status:      tx["tstatus"].(string),
			Doc:         doc,
			Postings:    postings,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
