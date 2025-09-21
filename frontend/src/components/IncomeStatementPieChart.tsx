"use client";

import * as React from "react";
import { DateRange } from "react-day-picker";
import {
  Pie,
  PieChart,
  Cell,
  Label,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  LabelList,
} from "recharts";

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
import { formatLocalDate } from "@/lib/utils";
import { Button } from "./ui/button";
import { BarChartBigIcon, CornerLeftUp, PieChartIcon } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "./ui/tooltip";
import { Tabs, TabsList, TabsTrigger } from "./ui/tabs";
import { TabsContent } from "@radix-ui/react-tabs";

type IncomeStatementData = {
  account: string;
  amount: number;
  currency: string;
};

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

type Total = { amount: number; currency: string };

// mock API (replace with your fetch)
async function getIncomeStatementData(
  startDate: string,
  endDate: string,
  account: string,
  depth: number
) {
  const res = await fetch(
    `http://localhost:8080/api/incomestatement/?outputFormat=json&startDate=${startDate}&endDate=${endDate}&account=${account}&depth=${depth}&valueMode=then`
  );
  const data = await res.json();
  return data;
}

type IncomeStatementPieChartProps = {
  range: DateRange | undefined;
  rootAccount?: string;
  title: string;
  description: string;
};

export function IncomeStatementPieChart({
  range,
  rootAccount,
  title,
  description,
}: IncomeStatementPieChartProps) {
  const [chartData, setChartData] = React.useState<IncomeStatementData[]>([]);
  const [chartConfig, setChartConfig] = React.useState<ChartConfig>({});
  const [total, setTotal] = React.useState<Total>({
    amount: 0,
    currency: "USD",
  });
  const [noData, setNoData] = React.useState(false);
  const [account, setAccount] = React.useState<string[]>([rootAccount || ""]);

  React.useEffect(() => {
    if (rootAccount) {
      setAccount([rootAccount]);
    }
  }, [rootAccount]);

  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getIncomeStatementData(
        formatLocalDate(range?.from),
        formatLocalDate(range?.to),
        account[account.length - 1],
        account.length + 1
      );
      setTotal({
        amount: data.total.amount,
        currency: data.total.currency,
      });
      if (data.total.amount === 0) {
        setNoData(true);
      } else {
        setNoData(false);
        setChartData(data.incomeData);
      }
    };
    fetchData();
  }, [range, account]);

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
      <Tabs defaultValue="pie">
        <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
          <div className="grid flex-1 gap-1">
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          {account.length > 1 && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="outline"
                  onClick={() => {
                    setAccount((prev) => prev.slice(0, -1));
                  }}
                >
                  <CornerLeftUp />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Go back to {account[account.length - 2]}</p>
              </TooltipContent>
            </Tooltip>
          )}
          <TabsList>
            <TabsTrigger value="pie">
              <PieChartIcon />
            </TabsTrigger>
            <TabsTrigger value="bar">
              <BarChartBigIcon />
            </TabsTrigger>
          </TabsList>
        </CardHeader>
        <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
          {noData && (
            <div className="flex h-[250px] w-full items-center justify-center">
              <p className="text-muted-foreground">
                No data for selected date range
              </p>
            </div>
          )}
          {!noData && (
            <>
              <TabsContent value="pie">
                <ChartContainer
                  config={chartConfig}
                  className="aspect-auto h-[250px] w-full"
                >
                  <PieChart>
                    <ChartTooltip
                      cursor={false}
                      content={<ChartTooltipContent />}
                    />
                    <Pie
                      data={chartData}
                      dataKey="amount"
                      nameKey="account"
                      innerRadius={50}
                      outerRadius={80}
                      onClick={(data) => {
                        if (account[account.length - 1] != data.account) {
                          setAccount((prev) => [...prev, data.account]);
                        }
                      }}
                    >
                      {chartData.map((entry) => (
                        <Cell
                          key={entry.account}
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
                                  {total.amount.toLocaleString("en-US", {
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
                                  {total.currency}
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
              </TabsContent>
              <TabsContent value="bar">
                <ChartContainer
                  config={chartConfig}
                  className="aspect-auto h-72 w-full"
                >
                  <BarChart
                    accessibilityLayer
                    data={chartData}
                    layout="vertical"
                  >
                    <ChartTooltip
                      cursor={false}
                      content={<ChartTooltipContent />}
                    />
                    <YAxis
                      dataKey="account"
                      type="category"
                      tickLine={false}
                      tickMargin={10}
                      axisLine={false}
                      hide
                    />
                    <XAxis
                      dataKey="amount"
                      type="number"
                      tickLine={false}
                      axisLine={false}
                      tickMargin={8}
                      tickFormatter={(value: number) =>
                        `${value.toLocaleString("en-US", {
                          maximumFractionDigits: 1,
                          notation: "compact",
                          compactDisplay: "short",
                        })} ${
                          chartData[Math.ceil(chartData.length / 2)]
                            ?.currency ||
                          chartData[0]?.currency ||
                          ""
                        }`
                      }
                    />
                    <Bar
                      dataKey="amount"
                      layout="vertical"
                      radius={5}
                      onClick={(data) => {
                        if (account[account.length - 1] != data.account) {
                          setAccount((prev) => [...prev, data.account]);
                        }
                      }}
                    >
                      <LabelList
                        dataKey="account"
                        position="insideLeft"
                        offset={8}
                        className="fill-accent"
                        fontSize={12}
                      />
                      {chartData.map((entry) => (
                        <Cell
                          key={entry.account}
                          fill={chartConfig[entry.account.toLowerCase()]?.color}
                        />
                      ))}
                    </Bar>
                  </BarChart>
                </ChartContainer>
                {/* <ChartLegend
                  className="flex-wrap"
                  content={<ChartLegendContent />}
                /> */}
              </TabsContent>
            </>
          )}
        </CardContent>
      </Tabs>
    </Card>
  );
}
