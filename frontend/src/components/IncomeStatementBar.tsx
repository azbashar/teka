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

import { Config, useConfig } from "@/context/ConfigContext";

async function getIncomeStatementData(startDate: string, endDate: string) {
  console.log(`from: ${startDate} to: ${endDate}`);
  const res = await fetch(
    `http://localhost:8080/api/incomestatement/?outputFormat=json&startDate=${startDate}&endDate=${endDate}&depth=1&valueMode=then&period=M`
  );
  const data = await res.json();
  return data;
}

type IncomeStatementBarProps = {
  range: DateRange | undefined;
  title: string;
  description: string;
};

type IncomeStatementData = {
  period: string;
  income: number;
  expense: number;
  currency: string;
}[];

type RawIncomeStatementData = {
  incomeData: {
    data: {
      account: string;
      amount: number;
      currency: string;
    }[];
    dates: {
      from: string;
      to: string;
    };
    total: {
      amount: number;
      currency: string;
    };
  }[];
};

function transformToChartData(
  rawData: RawIncomeStatementData,
  config: Config | null
): IncomeStatementData {
  return rawData.incomeData.map((period) => {
    const incomeItem = period.data.find(
      (d) => d.account == config?.Accounts.IncomeAccount
    );
    const expenseItem = period.data.find(
      (d) => d.account == config?.Accounts.ExpenseAccount
    );

    return {
      period: `${new Date(period.dates.from).toLocaleDateString("en-US", {
        month: "short",
        year: "2-digit",
      })}`,
      income: incomeItem?.amount || 0,
      expense: expenseItem?.amount || 0,
      currency: incomeItem?.currency || expenseItem?.currency || "",
    };
  });
}

export function IncomeStatementBar({
  range,
  title,
  description,
}: IncomeStatementBarProps) {
  const config = useConfig();

  const [chartData, setChartData] = React.useState<IncomeStatementData>([]);
  const [noData, setNoData] = React.useState(false);

  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getIncomeStatementData(
        formatLocalDate(
          new Date(
            range?.from?.getFullYear() ?? 1970,
            range?.from?.getMonth() ?? 1,
            1
          )
        ), // round date to start of month
        formatLocalDate(
          new Date(
            range?.to?.getFullYear() ?? 1970,
            (range?.to?.getMonth() ?? 1) + 1,
            0
          )
        ) // round date to end of month
      );

      if (!data.incomeData) {
        setNoData(true);
      } else {
        setNoData(false);
        setChartData(transformToChartData(data, config));
      }
    };
    fetchData();
  }, [range, config]);

  const chartConfig = {
    income: {
      label: "Income",
      color: "var(--chart-1)",
    },
    expense: {
      label: "Expenses",
      color: "var(--chart-2)",
    },
  } satisfies ChartConfig;

  return (
    <Card className="pt-0">
      <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
        <div className="grid flex-1 gap-1">
          <CardTitle>{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </div>
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
          <ChartContainer
            config={chartConfig}
            className="aspect-auto h-[250px] w-full"
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
                    compactDisplay: "short",
                  })} ${
                    chartData[Math.ceil(chartData.length / 2)]?.currency ||
                    chartData[0]?.currency ||
                    ""
                  }`
                }
                width={80}
              />
              <ChartTooltip
                cursor={false}
                content={<ChartTooltipContent indicator="dashed" />}
              />
              <Bar dataKey="income" fill="var(--color-income)" />
              <Bar dataKey="expense" fill="var(--color-expense)" />
              <ChartLegend content={<ChartLegendContent />} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
