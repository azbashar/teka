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

type IncomeData = { account: string; amount: number; currency: string };

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

type TotalIncome = { amount: number; currency: string };

// mock API (replace with your fetch)
async function getNetWorthData(startDate: string, endDate: string) {
  const res = await fetch(
    `http://localhost:8080/api/incomeDistribution/?startDate=${startDate}&endDate=${endDate}`
  );
  const data = await res.json();
  return data;
}

type IncomePieChartProps = {
  range: DateRange | undefined;
};

export function IncomePieChart({ range }: IncomePieChartProps) {
  const [chartData, setChartData] = React.useState<IncomeData[]>([]);
  const [chartConfig, setChartConfig] = React.useState<ChartConfig>({});
  const [totalIncome, setTotalIncome] = React.useState<TotalIncome>({
    amount: 0,
    currency: "USD",
  });
  const [noIncome, setNoIncome] = React.useState(false);

  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getNetWorthData(
        range?.from?.toISOString().slice(0, 10) || "",
        range?.to?.toISOString().slice(0, 10) || ""
      );
      setTotalIncome({
        amount: data.total.amount,
        currency: data.total.currency,
      });
      if (data.total.amount === 0) {
        setNoIncome(true);
      } else {
        setNoIncome(false);
        setChartData(data.incomeData);
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
          <CardTitle>Income</CardTitle>
          <CardDescription>Your income distribution</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        {noIncome && (
          <div className="flex h-[250px] w-full items-center justify-center">
            <p className="text-muted-foreground">
              No income during selected date range
            </p>
          </div>
        )}
        {!noIncome && (
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
                            {totalIncome.amount.toLocaleString("en-US", {
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
                            {totalIncome.currency}
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
