"use client";
import { DateRangePicker } from "@/components/DateRangePicker";
import { StatementBar } from "@/components/StatementBar";
import { StatementPieChart } from "@/components/StatementPieChart";
import { StatementStackedBar } from "@/components/StatementStackedBar";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Toggle } from "@/components/ui/toggle";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";

import { useConfig } from "@/context/ConfigContext";
import { usePageTitle } from "@/context/PageTitleContext";
import { formatLocalDate } from "@/lib/utils";
import { ScaleIcon } from "lucide-react";
import * as React from "react";
import { DateRange } from "react-day-picker";

// get income statement html from api localhost:800/api/incomestatement?startDate=range.from&endDate=range.to
async function getIncomeStatement(
  startDate: string,
  endDate: string,
  valueMode: string,
  period: string
) {
  const res = await fetch(
    `http://localhost:8080/api/incomestatement/?startDate=${startDate}&endDate=${endDate}&valueMode=${valueMode}&period=${period}&outputFormat=html`
  );
  let data = await res.text();

  data = data.replace(/<style[\s\S]*?<\/style>/gi, "");

  // Remove <link rel="stylesheet" ...> tags
  data = data.replace(/<link\s+rel=["']stylesheet["'][^>]*>/gi, "");

  data =
    `<div class="incomestatementdata">` +
    data +
    `</div> <style> 
  	.incomestatementdata tr:nth-child(2n+1) td {
    	background-color: var(--color-table-odd-row) !important;
	}
	.incomestatementdata tr:has(.coltotal) > td {
  		border-top: double var(--color-foreground) !important;
	}
	</style>`;
  return data;
}

export default function IncomeStatementPage() {
  const config = useConfig();
  const { setTitle } = usePageTitle();
  React.useEffect(() => {
    setTitle("Income Statement");
  }, [setTitle]);

  const [range, setRange] = React.useState<DateRange | undefined>({
    from: new Date(new Date().setFullYear(new Date().getFullYear() - 1)),
    to: new Date(),
  });

  const [valueMode, setValueMode] = React.useState(true);

  const [period, setPeriod] = React.useState("");

  const [incomeStatementHTML, setIncomeStatementHTML] = React.useState("");
  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getIncomeStatement(
        formatLocalDate(range?.from),
        formatLocalDate(
          range?.to
            ? new Date(range.to.getTime() + 24 * 60 * 60 * 1000)
            : new Date()
        ),
        valueMode ? "then" : "",
        period
      );
      setIncomeStatementHTML(data);
    };
    fetchData();
  }, [range, valueMode, period]);

  return (
    <div>
      <div className="mb-4">
        <div className="flex flex-col sm:flex-row-reverse justify-between gap-4 sm:items-center mb-2">
          <DateRangePicker range={range} onChange={setRange} />
          <div className="flex items-baseline gap-1">
            <h2 className="font-semibold text-xl">Income Statement</h2>
          </div>
        </div>
        <div className="flex gap-2 mb-2">
          <ToggleGroup
            aria-label="Select period"
            type="single"
            value={period}
            onValueChange={(value) => setPeriod(value)}
          >
            <ToggleGroupItem
              variant="outline"
              aria-label="Toggle period monthly"
              value="M"
            >
              Monthly
            </ToggleGroupItem>
            <ToggleGroupItem
              variant="outline"
              aria-label="Toggle period quarterly"
              value="Q"
            >
              Quarterly
            </ToggleGroupItem>
            <ToggleGroupItem
              variant="outline"
              aria-label="Toggle period yearly"
              value="Y"
            >
              Yearly
            </ToggleGroupItem>
          </ToggleGroup>
          <Toggle
            aria-label="Toggle value at local"
            variant="outline"
            onPressedChange={(pressed) => setValueMode(pressed)}
            pressed={valueMode}
          >
            <ScaleIcon />
            Value at Local
          </Toggle>
        </div>
        <div className="flex">
          <ScrollArea className="w-1 border rounded p-4 bg-sidebar flex-1">
            <div dangerouslySetInnerHTML={{ __html: incomeStatementHTML }} />
            <ScrollBar orientation="horizontal" />
          </ScrollArea>
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <StatementBar
          range={range}
          title="Income and Expenses Growth"
          description="Your income and expenses per month."
          statement="incomestatement"
        />

        <div className="grid lg:grid-cols-2 gap-4">
          <StatementStackedBar
            range={range}
            title="Income Growth"
            description="Your income per month by account."
            rootAccount={config?.Accounts.IncomeAccount}
            statement="incomestatement"
          />
          <StatementStackedBar
            range={range}
            title="Expenses Growth"
            description="Your expenses per month by account."
            rootAccount={config?.Accounts.ExpenseAccount}
            statement="incomestatement"
          />
        </div>
        <div className="grid lg:grid-cols-2 gap-4">
          <StatementPieChart
            range={range}
            title="Income"
            description="Your income distribution."
            rootAccount={config?.Accounts.IncomeAccount}
            statement="incomestatement"
          />
          <StatementPieChart
            range={range}
            title="Expenses"
            description="Your expense distribution."
            rootAccount={config?.Accounts.ExpenseAccount}
            statement="incomestatement"
          />
        </div>
      </div>
    </div>
  );
}
