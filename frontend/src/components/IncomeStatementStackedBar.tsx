"use client";

import * as React from "react";
import { DateRange } from "react-day-picker";
import { BarChart, Bar, XAxis, YAxis, CartesianGrid } from "recharts";

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
};

type RawIncomeStatementData = {
  incomeData: {
    data: { account: string; amount: number; currency: string }[];
    dates: { from: string; to: string };
    total: { amount: number; currency: string };
  }[];
};

async function getIncomeStatementData(
  startDate: string,
  endDate: string,
  account: string,
  depth: number
) {
  const res = await fetch(
    `http://localhost:8080/api/incomestatement/?outputFormat=json&startDate=${startDate}&endDate=${endDate}&depth=${depth}&valueMode=then&period=M&account=${account}`
  );
  return res.json();
}

export function IncomeStatementStackedBar({
  range,
  title,
  description,
  rootAccount,
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
    if (rootAccount) {
      setAccount([rootAccount]);
    }
  }, [rootAccount]);

  React.useEffect(() => {
    const fetchData = async () => {
      const data: RawIncomeStatementData = await getIncomeStatementData(
        formatLocalDate(
          new Date(
            range?.from?.getFullYear() ?? 1970,
            range?.from?.getMonth() ?? 0,
            1
          )
        ),
        formatLocalDate(
          new Date(
            range?.to?.getFullYear() ?? 1970,
            (range?.to?.getMonth() ?? 0) + 1,
            0
          )
        ),
        account[account.length - 1],
        account.length + 1
      );

      if (!data?.incomeData || data.incomeData.length === 0) {
        setNoData(true);
        return;
      }

      setNoData(false);
      setCurrency(data.incomeData[0].total.currency ?? "");

      // Collect all unique accounts
      const allAccounts = Array.from(
        new Set(data.incomeData.flatMap((p) => p.data.map((d) => d.account)))
      );
      setAccounts(allAccounts);

      const transformed = data.incomeData.map((period) => {
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
    };

    fetchData();
  }, [range, account, config]);

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
              <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
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
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
