"use client";

import * as React from "react";
import { DateRange } from "react-day-picker";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Brush } from "recharts";

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
import { useConfig } from "@/context/ConfigContext";
import { Tooltip, TooltipContent, TooltipTrigger } from "./ui/tooltip";
import { Button } from "./ui/button";
import { CornerLeftUp } from "lucide-react";
import { toast } from "sonner";

const colors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

type StackedBarChartProps = {
  range: DateRange | undefined;
  title: string;
  description: string;
  rootAccount?: string;
  statement: "incomestatement" | "balancesheet";
};

type RawStatementData = {
  data: { account: string; amount: number; currency: string }[];
  dates: { from: string; to: string };
  total: { amount: number; currency: string };
}[];
export function StatementStackedBar({
  range,
  title,
  description,
  rootAccount,
  statement,
}: StackedBarChartProps) {
  const config = useConfig();
  const [chartData, setChartData] = React.useState<
    Record<string, number | string>[]
  >([]);
  const [accounts, setAccounts] = React.useState<string[]>([]);
  const [currency, setCurrency] = React.useState<string>("");
  const [noData, setNoData] = React.useState(false);
  const [account, setAccount] = React.useState<string[]>([rootAccount || ""]);

  React.useEffect(() => {
    const fetchData = async () => {
      const startDate = formatLocalDate(
        new Date(
          range?.from?.getFullYear() ?? 1970,
          range?.from?.getMonth() ?? 0,
          1
        )
      );
      const endDate = formatLocalDate(
        new Date(
          range?.to?.getFullYear() ?? 1970,
          (range?.to?.getMonth() ?? 0) + 1,
          0
        )
      );
      const acct = account[account.length - 1];
      const depth = account.length + 1;
      const value = statement == "incomestatement" ? "then" : "end";
      fetch(
        `http://localhost:8080/api/${statement}/?outputFormat=json&startDate=${startDate}&endDate=${endDate}&depth=${depth}&valueMode=${value}&period=M&account=${acct}`
      )
        .then((res) => {
          if (!res.ok) {
            return res.text().then((body) => {
              throw new Error(`(${res.status}) ${res.statusText} : ${body}}`);
            });
          }
          return res.json();
        })
        .then((data: RawStatementData) => {
          if (!data || data.length === 0) {
            setNoData(true);
            return;
          }

          setNoData(false);
          setCurrency(data[0].total.currency ?? "");

          // Collect all unique accounts
          const allAccounts = Array.from(
            new Set(data.flatMap((p) => p.data.map((d) => d.account)))
          );
          setAccounts(allAccounts);

          const transformed = data.map((period) => {
            const base: Record<string, number | string> = {
              period: new Date(period.dates.from).toLocaleDateString("en-US", {
                month: "short",
                year: "2-digit",
              }),
            };
            allAccounts.forEach((account) => {
              const item = period.data.find((d) => d.account === account);
              base[account] = item?.amount ?? 0;
            });
            return base;
          });

          setChartData(transformed);
        })
        .catch((err) => {
          toast.error(`Error fetching data: ${err.message}`);
          console.error(
            `Component: StatementStackedBar, Error fetching data: ${err.message}`
          );
        });
    };

    fetchData();
  }, [range, account, config, statement]);

  // Build chartConfig dynamically for shadcn legend/colors
  const chartConfig: ChartConfig = accounts.reduce((acc, account, idx) => {
    acc[account] = {
      label: account,
      color: colors[idx % colors.length],
    };
    return acc;
  }, {} as ChartConfig);

  return (
    <Card className="pt-0">
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
                onClick={() => setAccount((prev) => prev.slice(0, -1))}
              >
                <CornerLeftUp />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Go back to {account[account.length - 2]}</p>
            </TooltipContent>
          </Tooltip>
        )}
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        {noData ? (
          <div className="flex h-[250px] w-full items-center justify-center">
            <p className="text-muted-foreground">
              No data for selected date range
            </p>
          </div>
        ) : (
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[300px] w-full"
          >
            <BarChart accessibilityLayer data={chartData}>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="period"
                type="category"
                tickLine={false}
                axisLine={false}
                height={40}
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                tickMargin={8}
                tickFormatter={(value: number) =>
                  `${value.toLocaleString("en-US", {
                    maximumFractionDigits: 1,
                    notation: "compact",
                  })} ${currency}`
                }
                width={80}
              />
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent />}
                formatter={(value, name, item) => (
                  <div className="flex items-center gap-2 w-full">
                    <div
                      className="w-2 h-2 rounded-xs"
                      style={{ backgroundColor: item.color }}
                    ></div>
                    <span className="flex gap-2 justify-between flex-1">
                      {name}{" "}
                      <span className="text-muted-foreground">
                        {value.toLocaleString("en-US")} {currency}
                      </span>
                    </span>
                  </div>
                )}
              />
              {accounts.map((acct) => (
                <Bar
                  key={acct}
                  dataKey={acct}
                  stackId="a"
                  fill={chartConfig[acct]?.color}
                  onClick={() => {
                    if (account[account.length - 1] !== acct) {
                      setAccount((prev) => [...prev, acct]);
                    }
                  }}
                />
              ))}
              <ChartLegend
                className="flex-wrap"
                content={<ChartLegendContent />}
              />
              <Brush
                dataKey="period"
                height={10}
                stroke="var(--color-accent)"
                fill="var(--color-background)"
              />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
