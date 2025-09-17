"use client";
import { DateRangePicker } from "@/components/DateRangePicker";
import { ExpensePieChart } from "@/components/ExpensePieChart";
import { IncomePieChart } from "@/components/IncomePieChart";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Toggle } from "@/components/ui/toggle";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { usePageTitle } from "@/context/PageTitleContext";
import { formatLocalDate } from "@/lib/utils";
import { HelpCircle, ScaleIcon } from "lucide-react";
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
      <title>Income Statement</title>
      <div className="mb-4">
        <div className="flex flex-col sm:flex-row-reverse justify-between gap-4 sm:items-center mb-2">
          <DateRangePicker range={range} onChange={setRange} />
          <div className="flex items-baseline gap-1">
            <h2 className="font-semibold text-xl">Income Statement</h2>
            <Tooltip>
              <TooltipTrigger>
                <HelpCircle className="size-3 text-gray-200" />
              </TooltipTrigger>
              <TooltipContent>
                <p>
                  Accounts added to <code>starred_accounts</code> in the config
                  will be shown here.
                </p>
              </TooltipContent>
            </Tooltip>
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
        <div className="grid lg:grid-cols-2 gap-4">
          <IncomePieChart range={range} />
          <ExpensePieChart range={range} />
        </div>
      </div>
    </div>
  );
}
