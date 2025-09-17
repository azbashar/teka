"use client";

import * as React from "react";
import { DateRange } from "react-day-picker";
import { Pie, PieChart, Cell, Label } from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";

import { DateRangePicker } from "./DateRangePicker";
import { formatLocalDate } from "@/lib/utils";

type ExpenseData = { account: string; amount: number; currency: string };

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

type TotalExpense = { amount: number; currency: string };

// mock API (replace with your fetch)
async function getNetWorthData(startDate: string, endDate: string) {
  const res = await fetch(
    `http://localhost:8080/api/expenseDistribution/?startDate=${startDate}&endDate=${endDate}`
  );
  const data = await res.json();
  return data;
}

type ExpensePieChartProps = {
  range: DateRange | undefined;
};

export function ExpensePieChart({ range }: ExpensePieChartProps) {
  const [chartData, setChartData] = React.useState<ExpenseData[]>([]);
  const [chartConfig, setChartConfig] = React.useState<ChartConfig>({});
  const [totalExpense, setTotalExpense] = React.useState<TotalExpense>({
    amount: 0,
    currency: "USD",
  });
  const [noExpense, setNoExpense] = React.useState(false);

  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getNetWorthData(
        formatLocalDate(range?.from),
        formatLocalDate(range?.to)
      );
      setTotalExpense({
        amount: data.total.amount,
        currency: data.total.currency,
      });
      if (data.total.amount === 0) {
        setNoExpense(true);
      } else {
        setNoExpense(false);
        setChartData(data.expenseData);
      }
    };
    fetchData();
  }, [range]);

  // rebuild chartConfig whenever chartData changes
  React.useEffect(() => {
    const config = Object.fromEntries(
      chartData.map((d, i) => [
        d.account.toLowerCase(),
        { label: d.account, color: colors[i % colors.length] },
      ])
    ) satisfies ChartConfig;
    setChartConfig(config);
  }, [chartData]);

  return (
    <Card className="pt-0">
      <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
        <div className="grid flex-1 gap-1">
          <CardTitle>Expenses</CardTitle>
          <CardDescription>Your expense distribution</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        {noExpense && (
          <div className="flex h-[250px] w-full items-center justify-center">
            <p className="text-muted-foreground">
              No Expense during selected date range
            </p>
          </div>
        )}
        {!noExpense && (
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[250px] w-full"
          >
            <PieChart>
              <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
              <Pie
                data={chartData}
                dataKey="amount"
                nameKey="account"
                innerRadius={50}
                outerRadius={80}
              >
                {chartData.map((entry, i) => (
                  <Cell
                    key={`cell-${i}`}
                    fill={chartConfig[entry.account.toLowerCase()]?.color}
                  />
                ))}
                <Label
                  content={({ viewBox }) => {
                    if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                      return (
                        <text
                          x={viewBox.cx}
                          y={viewBox.cy}
                          textAnchor="middle"
                          dominantBaseline="middle"
                        >
                          <tspan
                            x={viewBox.cx}
                            y={viewBox.cy}
                            className="fill-foreground text-3xl font-bold"
                          >
                            {totalExpense.amount.toLocaleString("en-US", {
                              maximumFractionDigits: 1,
                              notation: "compact",
                              compactDisplay: "short",
                            })}
                          </tspan>
                          <tspan
                            x={viewBox.cx}
                            y={(viewBox.cy || 0) + 24}
                            className="fill-muted-foreground"
                          >
                            {totalExpense.currency}
                          </tspan>
                        </text>
                      );
                    }
                  }}
                />
              </Pie>
              <ChartLegend
                className="flex-wrap"
                content={<ChartLegendContent />}
              />
            </PieChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
