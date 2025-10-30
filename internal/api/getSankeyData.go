// WARNING: Entirely AI generated slop but some how seems to work.
package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/azbashar/teka/internal/config"
	"github.com/azbashar/teka/internal/fileselector"
)

type SankeyNode struct {
	Name string `json:"name"`
}

type SankeyLink struct {
	Source int     `json:"source"`
	Target int     `json:"target"`
	Value  float64 `json:"value"`
}

type SankeyResponse struct {
	Nodes []SankeyNode `json:"nodes"`
	Links []SankeyLink `json:"links"`
}

// Helper function to get Net Income (Remains Unchanged)
func getNetIncome(isData map[string]any) (float64, error) {
	// ... (Implementation for getNetIncome remains the same) ...
	getAmount := func(arr any, i int) float64 {
		if arrList, ok := arr.([]any); ok && len(arrList) > i {
			if inner, ok := arrList[i].([]any); ok && len(inner) > 0 {
				if amtMap, ok := inner[0].(map[string]any); ok {
					if aq, ok := amtMap["aquantity"].(map[string]any); ok {
						if fp, ok := aq["floatingPoint"].(float64); ok {
							return fp
						}
					}
				}
			}
		}
		return 0
	}

	cbrTotals, ok := isData["cbrTotals"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("cbrTotals not found in income statement data")
	}

	netIncome := getAmount(cbrTotals["prrAmounts"], 0)
	if netIncome != 0 {
		return netIncome, nil
	}
	return 0, fmt.Errorf("net Income amount not found or invalid format")
}

func getSankeyData(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// ... (Parameter parsing and hledger command execution code remains the same) ...
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	depthStr := r.URL.Query().Get("depth")
	depth := 1
	if depthStr != "" {
		var err error
		depth, err = strconv.Atoi(depthStr)
		if err != nil || depth < 1 {
			http.Error(w, "Invalid depth value. It should be a positive integer.", http.StatusBadRequest)
			return
		}
	}

	files, expr, err := fileselector.GetRequiredFiles(startDate, endDate, fileArg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ----- HLedger Income/Expense -----
	isArgs := []string{"is", "-O", "json", "--value=then," + config.Cfg.BaseCurrency, "--cost", "expr: not tag:clopen"}
	if depthStr != "" {
		isArgs = append(isArgs, fmt.Sprintf("--depth=%s", depthStr))
	}
	for _, f := range files {
		isArgs = append(isArgs, "-f", f)
	}
	if expr != "" {
		isArgs = append(isArgs, expr)
	}
	if startDate != "" {
		isArgs = append(isArgs, "-b", startDate)
	}
	if endDate != "" {
		isArgs = append(isArgs, "-e", endDate)
	}

	isOut, err := exec.Command("hledger", isArgs...).CombinedOutput()
	if err != nil {
		fmt.Println(string(isOut))
		http.Error(w, fmt.Sprintf("hledger income statement error: %v", err), http.StatusInternalServerError)
		return
	}

	var isData map[string]any
	if err := json.Unmarshal(isOut, &isData); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse income statement: %v", err), http.StatusInternalServerError)
		return
	}

	// ----- HLedger Balance Sheet -----
	bsArgs := []string{"bs", "--change", "-O", "json", "--value=then," + config.Cfg.BaseCurrency, "--cost", "expr: not tag:clopen"}
	if depthStr != "" {
		bsArgs = append(bsArgs, fmt.Sprintf("--depth=%s", depthStr))
	}
	for _, f := range files {
		bsArgs = append(bsArgs, "-f", f)
	}
	if expr != "" {
		bsArgs = append(bsArgs, expr)
	}
	if startDate != "" {
		bsArgs = append(bsArgs, "-b", startDate)
	}
	if endDate != "" {
		bsArgs = append(bsArgs, "-e", endDate)
	}

	bsOut, err := exec.Command("hledger", bsArgs...).CombinedOutput()
	if err != nil {
		fmt.Println(string(bsOut))
		http.Error(w, fmt.Sprintf("hledger balance sheet error: %v", err), http.StatusInternalServerError)
		return
	}

	var bsData map[string]any
	if err := json.Unmarshal(bsOut, &bsData); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse balance sheet: %v", err), http.StatusInternalServerError)
		return
	}

	// ----- Build Nodes & Links -----
	nodes := []SankeyNode{}
	nodeIndex := map[string]int{}
	linkMap := map[string]float64{}

	// Map to track split nodes: Key="AccountName-Direction", Value=NodeIndex
	// Direction: "Source" (Child->Parent) or "Target" (Parent->Child)
	splitNodeIndex := map[string]int{}

	// Map to track the flow through the split nodes to create the internal link (Source -> Target)
	// Key: "ParentName-Direction", Value: Total flow
	splitFlowTotals := map[string]float64{}

	// 1. Helper for Leaf Nodes and Top Roots (No suffix)
	addNode := func(name string) int {
		if idx, ok := nodeIndex[name]; ok {
			return idx
		}
		idx := len(nodes)
		nodes = append(nodes, SankeyNode{Name: name})
		nodeIndex[name] = idx
		return idx
	}

	// 2. Helper for Intermediate Parent Nodes (with suffix)
	// isSource: true if this node is acting as a funder (Source), false if it's acting as a funded (Target)
	getSplitNodeIndex := func(name string, isSource bool) int {
		var dirSuffix string
		if isSource {
			dirSuffix = " " // Receiving flow UP from child
		} else {
			dirSuffix = "" // Sending flow DOWN to child
		}

		key := name + dirSuffix

		if idx, ok := splitNodeIndex[key]; ok {
			return idx
		}

		// Create the node
		idx := len(nodes)
		nodes = append(nodes, SankeyNode{Name: key})
		splitNodeIndex[key] = idx

		return idx
	}

	// Helper to add link to map
	addLink := func(source, target int, value float64) {
		if value > 0.001 { // Only add links with a meaningful positive value
			key := fmt.Sprintf("%d-%d", source, target)
			linkMap[key] += value
		}
	}

	// Helper to get floatingPoint amount safely (Remains Unchanged)
	getAmount := func(arr any, i int) float64 {
		// ... (Implementation for getAmount remains the same) ...
		if arrList, ok := arr.([]any); ok && len(arrList) > i {
			if inner, ok := arrList[i].([]any); ok && len(inner) > 0 {
				if amtMap, ok := inner[0].(map[string]any); ok {
					if aq, ok := amtMap["aquantity"].(map[string]any); ok {
						if fp, ok := aq["floatingPoint"].(float64); ok {
							return fp
						}
					}
				}
			}
		}
		return 0
	}

	// --- Recursive hierarchy builder with SPLIT NODE LOGIC ---
	// This function is now ONLY used for Income/Expenses, which don't typically have conflicting flows.
	// We'll keep the original simple hierarchy logic for I/E for simplicity, but adjust it slightly.
	addRowWithHierarchy := func(fullName string, value float64, direction bool) {
		parts := strings.Split(fullName, ":")
		var currentPath string

		absValue := math.Abs(value)
		if absValue < 0.001 {
			return
		}

		for i, part := range parts {
			if i == 0 {
				currentPath = part
				continue
			}

			parentPath := currentPath
			currentPath += ":" + part

			parentIdx := addNode(parentPath)
			childIdx := addNode(currentPath)

			var source, target int
			if direction {
				// Parent -> Child (Use of Funds)
				source, target = parentIdx, childIdx
			} else {
				// Child -> Parent (Source of Funds)
				source, target = childIdx, parentIdx
			}

			// Use the absolute value for the flow quantity
			addLink(source, target, absValue)
		}

		addNode(fullName)
	}

	// --- New Hierarchy Builder for Balance Sheet with Mandatory Splitting ---
	// This function is specialized for Assets/Liabilities where flows conflict
	// flowDirection: true = Parent -> Child (Use of Funds), false = Child -> Parent (Source of Funds)
	addBalanceSheetHierarchy := func(fullName string, value float64, flowDirection bool) {
		parts := strings.Split(fullName, ":")
		var currentPath string

		absValue := math.Abs(value)
		if absValue < 0.001 {
			return
		}

		for i, part := range parts {
			if i == 0 {
				currentPath = part
				continue
			}

			parentPath := currentPath
			currentPath += ":" + part

			var sourceIdx, targetIdx int

			// Determine if the current node is a leaf (doesn't have children in the report at this depth)
			isLeaf := i == len(parts)-1

			// 1. Get the Child Index
			var childIdx int
			if isLeaf {
				childIdx = addNode(currentPath) // Leaf is a simple node
			} else {
				// Intermediate child is also a split node (gets flow from parent, passes flow to its own children)
				// It needs a single split node to handle the flow from the parent at this level
				childIdx = getSplitNodeIndex(currentPath, !flowDirection) // Child receives flow from parent, so it's the opposite of the current flowDirection
			}

			// 2. Get the Parent Index (ALWAYS a split node at the intermediate level)
			// For Assets, "assets" is the root. "assets:bank" is the intermediate parent.
			var parentIdx int
			if i == 1 {
				parentIdx = addNode(parentPath) // The ROOT ("assets") is NOT split here, it is linked later
			} else {
				// The intermediate parent is a split node.
				// If flowDirection is true (Parent->Child), Parent is the Target side.
				// If flowDirection is false (Child->Parent), Parent is the Source side.
				parentIdx = getSplitNodeIndex(parentPath, !flowDirection)
			}

			// 3. Create the Link
			if flowDirection {
				// Parent -> Child (Use of Funds)
				// Link Parent's [Target] side to Child's [Target] side (or Leaf node)
				sourceIdx = parentIdx
				targetIdx = childIdx
			} else {
				// Child -> Parent (Source of Funds)
				// Link Leaf node to Parent's [Source] side
				sourceIdx = childIdx
				targetIdx = parentIdx
			}

			addLink(sourceIdx, targetIdx, absValue)

			// 4. Track flow for internal split node linking later
			if i > 0 { // Only track flow for intermediate nodes, not the top root
				// The flow is always registered on the 'receiving' side of the parent/child connection
				// For flowDirection=true (P->C), the Child is receiving the flow.
				// For flowDirection=false (C->P), the Parent is receiving the flow.

				// Track the flow received by the parent's split node
				if i > 1 {
					parentKey := fmt.Sprintf("%s-%t", parentPath, !flowDirection)
					splitFlowTotals[parentKey] += absValue
				}

				// Track the flow received by the child's split node
				if !isLeaf {
					childKey := fmt.Sprintf("%s-%t", currentPath, flowDirection)
					splitFlowTotals[childKey] += absValue
				}
			}
		}

		addNode(fullName) // Ensure the leaf node is in the list
	}

	// --- Income/Expenses Processing (Uses Simple Hierarchy Builder) ---
	var incomeRows, expenseRows []any
	var rootIncomeName, rootExpenseName string
	expenseTotal := 0.0
	var incomeTotal float64 = 0.0

	// ... (I/E data extraction and linking remains the same as provided) ...
	if subs, ok := isData["cbrSubreports"].([]any); ok {
		for _, sub := range subs {
			subArr, _ := sub.([]any)
			if len(subArr) < 2 {
				continue
			}
			name := subArr[0].(string)
			data := subArr[1].(map[string]any)
			rows := data["prRows"].([]any)
			switch name {
			case "Revenues":
				incomeRows = rows
				if len(rows) > 0 {
					rootIncomeName = strings.Split(rows[0].(map[string]any)["prrName"].(string), ":")[0]
				}
				if prTotals, ok := data["prTotals"].(map[string]any); ok {
					incomeTotal = getAmount(prTotals["prrAmounts"], 0)
				}
			case "Expenses":
				expenseRows = rows
				if len(rows) > 0 {
					rootExpenseName = strings.Split(rows[0].(map[string]any)["prrName"].(string), ":")[0]
				}
				if prTotals, ok := data["prTotals"].(map[string]any); ok {
					expenseTotal = getAmount(prTotals["prrAmounts"], 0)
				}
			}
		}

		for _, row := range incomeRows {
			r := row.(map[string]any)
			name := r["prrName"].(string)
			val := getAmount(r["prrAmounts"], 0)
			addRowWithHierarchy(name, val, false) // Child -> Parent
		}
		for _, row := range expenseRows {
			r := row.(map[string]any)
			name := r["prrName"].(string)
			val := getAmount(r["prrAmounts"], 0)
			addRowWithHierarchy(name, val, true) // Parent -> Child
		}
		if len(incomeRows) > 0 && len(expenseRows) > 0 {
			rootIncomeIdx := addNode(rootIncomeName)
			rootExpenseIdx := addNode(rootExpenseName)
			flowValue := incomeTotal
			if incomeTotal > expenseTotal {
				flowValue = expenseTotal
			}
			addLink(rootIncomeIdx, rootExpenseIdx, flowValue)
		}
	}

	// --- Assets/Liabilities Processing (Uses Split Node Hierarchy Builder) ---
	var rootLiabilityName, rootAssetName string
	var totalLiabilityChange float64 = 0.0

	if subs, ok := bsData["cbrSubreports"].([]any); ok {
		for _, sub := range subs {
			subArr, _ := sub.([]any)
			if len(subArr) < 2 {
				continue
			}
			name := subArr[0].(string)
			data := subArr[1].(map[string]any)
			rows := data["prRows"].([]any)

			switch name {
			case "Liabilities":
				if len(rows) > 0 {
					rootLiabilityName = strings.Split(rows[0].(map[string]any)["prrName"].(string), ":")[0]
				}
			case "Assets":
				if len(rows) > 0 {
					rootAssetName = strings.Split(rows[0].(map[string]any)["prrName"].(string), ":")[0]
				}
			}

			for _, row := range rows {
				r := row.(map[string]any)
				name := r["prrName"].(string)
				val := getAmount(r["prrAmounts"], 0)

				if name == rootAssetName || name == rootLiabilityName {
					addNode(name)
					continue
				}

				var direction bool
				if strings.HasPrefix(name, rootAssetName) && rootAssetName != "" {
					direction = val >= 0 // ASSETS: Positive val (increase) is use of funds (true), Negative val (decrease) is source of funds (false).
				} else if strings.HasPrefix(name, rootLiabilityName) && rootLiabilityName != "" {
					direction = val < 0 // LIABILITIES: Positive val (increase) is source of funds (false), Negative val (decrease) is use of funds (true).
				} else {
					direction = true
				}

				addBalanceSheetHierarchy(name, val, direction)

				if name != rootLiabilityName && rootLiabilityName != "" && strings.HasPrefix(name, rootLiabilityName) {
					totalLiabilityChange += val
				}
			}
		}
	}

	// --- Internal Split Node Links (Crucial Step: Source -> Target) ---
	// Link the source side to the target side for every parent that received flows
	for key := range splitNodeIndex {
		parentName := strings.Split(key, " [")[0]

		sourceKey := fmt.Sprintf("%s [Source]", parentName)
		targetKey := fmt.Sprintf("%s [Target]", parentName)

		sourceIdx, sourceExists := splitNodeIndex[sourceKey]
		targetIdx, targetExists := splitNodeIndex[targetKey]

		sourceFlow := splitFlowTotals[sourceKey]
		targetFlow := splitFlowTotals[targetKey]

		if sourceExists && targetExists {
			// Link the net movement flow: from the side with the higher flow to the other side.
			// This represents the internal transfer between the source and target flows.
			// The value should be the MIN of the two flows to show how much *passes through* the account.
			// However, for Sankey visualization that needs to break the cycle, linking the higher flow to the lower
			// flow (or just linking the largest flow) helps visualize the primary direction.
			// For simplicity and cycle-breaking, let's link the Source (inflow) to the Target (outflow).
			// The value should be the total flow that passes through the account (or the minimum conflicting flow).

			// We can simplify this: link the "Source" side to the "Target" side with the total flow of the Source side
			// or the Target side, or the average. The simplest way to show a net effect and break recursion is
			// to link the source side to the target side with the larger flow.

			// Let's use the MINIMUM to represent the flow that *must* pass through this intermediate account to satisfy both sides.
			flow := math.Min(sourceFlow, targetFlow)
			if flow > 0.001 {
				addLink(sourceIdx, targetIdx, flow)
			}
		}
	}

	// --- Connect Top Level Roots ---
	netIncome, err := getNetIncome(isData)
	assetRootIdx := addNode("assets")
	incomeRootIdx := addNode(rootIncomeName)
	expenseRootIdx := addNode(rootExpenseName)

	if err == nil {
		absNetIncome := math.Abs(netIncome)
		if netIncome > 0 {
			addLink(incomeRootIdx, assetRootIdx, absNetIncome)
		} else if netIncome < 0 {
			addLink(assetRootIdx, expenseRootIdx, absNetIncome)
		}
	}

	liabilityRootIdx := addNode(rootLiabilityName)

	if totalLiabilityChange != 0 && rootLiabilityName != "" {
		absTotalLiabilityChange := math.Abs(totalLiabilityChange)
		if totalLiabilityChange > 0 {
			addLink(liabilityRootIdx, assetRootIdx, absTotalLiabilityChange)
		} else if totalLiabilityChange < 0 {
			addLink(assetRootIdx, liabilityRootIdx, absTotalLiabilityChange)
		}
	}

	// --- Link Split Sub-Roots to Main Root ---
	// This links the net flow from the split sub-account back to the main root (Assets/Liabilities).
	for name := range nodeIndex {
		if strings.Contains(name, ":") && strings.Count(name, ":") == 1 {
			parentRootName := strings.Split(name, ":")[0]

			if parentRootName == rootAssetName || parentRootName == rootLiabilityName {

				flowKeySource := fmt.Sprintf("%s [Source]", name)
				flowKeyTarget := fmt.Sprintf("%s [Target]", name)
				sourceFlow := splitFlowTotals[flowKeySource]
				targetFlow := splitFlowTotals[flowKeyTarget]

				// Net flow into the root account
				netFlow := sourceFlow - targetFlow // Positive netFlow means net source (Child -> Root)
				if math.Abs(netFlow) < 0.001 {
					continue
				}

				rootIdx := addNode(parentRootName)

				switch parentRootName {
				case rootAssetName:
					// Assets: Net Source (positive netFlow) means asset reduction, flows UP to root
					if netFlow > 0 {
						addLink(getSplitNodeIndex(name, true), rootIdx, netFlow)
					} else {
						// Net Use (negative netFlow) means asset increase, flows DOWN from root
						addLink(rootIdx, getSplitNodeIndex(name, false), math.Abs(netFlow))
					}
				case rootLiabilityName:
					// Liabilities: Net Source (positive netFlow) means liability increase, flows UP to root
					if netFlow > 0 {
						addLink(getSplitNodeIndex(name, true), rootIdx, netFlow)
					} else {
						// Net Use (negative netFlow) means liability reduction, flows DOWN from root
						addLink(rootIdx, getSplitNodeIndex(name, false), math.Abs(netFlow))
					}
				}
			}
		}
	}

	// --- Convert the aggregated linkMap back to the links slice ---
	links := []SankeyLink{}
	incoming := make(map[int]int) // node -> number of incoming edges
	outgoing := make(map[int]int) // node -> number of outgoing edges
	for key, val := range linkMap {
		parts := strings.Split(key, "-")
		source, _ := strconv.Atoi(parts[0])
		target, _ := strconv.Atoi(parts[1])
		links = append(links, SankeyLink{Source: source, Target: target, Value: val})
		outgoing[source]++
		incoming[target]++
	}

	// --- Compute Graph Metrics ---

	// totalSources: nodes with no incoming edges
	totalSources := 0
	// totalTargets: nodes with no outgoing edges
	totalTargets := 0
	for i := range nodes {
		if incoming[i] == 0 {
			totalSources++
		}
		if outgoing[i] == 0 {
			totalTargets++
		}
	}

	// maxChainLength: longest path in DAG-like Sankey graph
	var maxChainLength int
	var dfs func(node int, visited map[int]bool) int
	dfs = func(node int, visited map[int]bool) int {
		if visited[node] {
			return 0
		}
		visited[node] = true
		maxLen := 1
		for _, l := range links {
			if l.Source == node {
				length := 1 + dfs(l.Target, visited)
				if length > maxLen {
					maxLen = length
				}
			}
		}
		visited[node] = false
		return maxLen
	}

	for i := range nodes {
		if incoming[i] == 0 { // only start from sources
			length := dfs(i, map[int]bool{})
			if length > maxChainLength {
				maxChainLength = length
			}
		}
	}

	// currency from data
	currency := "" // Default currency
	if len(isData["cbrTotals"].(map[string]any)["prrAmounts"].([]any)) > 0 {
		if amtData, ok := isData["cbrTotals"].(map[string]any)["prrAmounts"].([]any)[0].([]any)[0].(map[string]any); ok {
			if comm, ok := amtData["acommodity"].(string); ok {
				currency = comm
			}
		}
	}

	// --- shorten node names (remove parent chain)
	for i := range nodes {
		parts := strings.Split(nodes[i].Name, ":")
		nodes[i].Name = parts[len(parts)-1]
	}

	// --- Wrap Response ---
	resp := map[string]any{
		"maxChainLength": maxChainLength,
		"totalSources":   totalSources,
		"totalTargets":   totalTargets,
		"currency":       currency,
		"sankeyData": SankeyResponse{
			Nodes: nodes,
			Links: links,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
