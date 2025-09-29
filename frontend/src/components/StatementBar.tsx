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

import { Config, useConfig } from "@/context/ConfigContext";
import { toast } from "sonner";

type IncomeStatementBarProps = {
  range: DateRange | undefined;
  title: string;
  description: string;
  statement: "incomestatement" | "balancesheet";
};

type ChartDataItem = {
  period: string;
  currency: string;
  // dynamic: will hold either income/expense or assets/liabilities
  [key: string]: string | number;
};

type RawIncomeStatementData = {
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

function transformToChartData(
  rawData: RawIncomeStatementData,
  config: Config | null,
  statement: "incomestatement" | "balancesheet"
): ChartDataItem[] {
  return rawData.map((period) => {
    const base = {
      period: `${new Date(period.dates.from).toLocaleDateString(
        config?.Locale,
        {
          month: "short",
          year: "2-digit",
        }
      )}`,
    };

    if (statement === "incomestatement") {
      const incomeItem = period.data.find(
        (d) => d.account === config?.Accounts.IncomeAccount
      );
      const expenseItem = period.data.find(
        (d) => d.account === config?.Accounts.ExpenseAccount
      );

      return {
        ...base,
        income: incomeItem?.amount || 0,
        expense: expenseItem?.amount || 0,
        currency: incomeItem?.currency || expenseItem?.currency || "",
      };
    } else {
      const assetsItem = period.data.find(
        (d) => d.account === config?.Accounts.AssetsAccount
      );
      const liabilitiesItem = period.data.find(
        (d) => d.account === config?.Accounts.LiabilitiesAccount
      );

      return {
        ...base,
        assets: assetsItem?.amount || 0,
        liabilities: liabilitiesItem?.amount || 0,
        currency: assetsItem?.currency || liabilitiesItem?.currency || "",
      };
    }
  });
}

export function StatementBar({
  range,
  title,
  description,
  statement,
}: IncomeStatementBarProps) {
  const config = useConfig();

  const [chartData, setChartData] = React.useState<ChartDataItem[]>([]);
  const [noData, setNoData] = React.useState(false);

  React.useEffect(() => {
    const fetchData = async () => {
      const startDate = formatLocalDate(
        new Date(
          range?.from?.getFullYear() ?? 1970,
          range?.from?.getMonth() ?? 1,
          1
        )
      ); // round date to start of month
      const endDate = formatLocalDate(
        new Date(
          range?.to?.getFullYear() ?? 1970,
          (range?.to?.getMonth() ?? 1) + 1,
          0
        )
      ); // round date to end of month
      const value = statement === "incomestatement" ? "then" : "end";

      fetch(
        `http://localhost:8080/api/${statement}/?outputFormat=json&startDate=${startDate}&endDate=${endDate}&depth=1&valueMode=${value}&period=M`
      )
        .then((res) => {
          if (!res.ok) {
            return res.text().then((body) => {
              throw new Error(`(${res.status}) ${res.statusText} : ${body}}`);
            });
          }
          return res.json();
        })
        .then((data) => {
          if (!data) {
            setNoData(true);
          } else {
            setNoData(false);
            setChartData(transformToChartData(data, config, statement));
          }
        })
        .catch((err) => {
          toast.error(`Error fetching data: ${err.message}`);
          console.error(
            `Component: StatementBar, Error fetching ${statement} data: ${err.message}`
          );
        });
    };
    fetchData();
  }, [range, config, statement]);

  // Dynamically build chart config depending on statement
  const chartConfig: ChartConfig =
    statement === "incomestatement"
      ? {
          income: {
            label: "Income",
            color: "var(--chart-1)",
          },
          expense: {
            label: "Expenses",
            color: "var(--chart-2)",
          },
        }
      : {
          assets: {
            label: "Assets",
            color: "var(--chart-1)",
          },
          liabilities: {
            label: "Liabilities",
            color: "var(--chart-2)",
          },
        };

  const keys = Object.keys(chartConfig); // ["income","expense"] or ["assets","liabilities"]

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
                  `${value.toLocaleString(config?.Locale, {
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
                formatter={(value, name, item) => (
                  <div className="flex items-center gap-2 w-full">
                    <div
                      className="w-2 h-2 rounded-xs"
                      style={{ backgroundColor: item.color }}
                    ></div>
                    <span className="flex gap-2 justify-between flex-1">
                      {chartConfig[name as keyof typeof chartConfig]?.label}{" "}
                      <span className="text-muted-foreground">
                        {Number(value).toLocaleString(config?.Locale)}{" "}
                        {chartData[Math.ceil(chartData.length / 2)]?.currency ||
                          chartData[0]?.currency ||
                          ""}
                      </span>
                    </span>
                  </div>
                )}
              />
              {keys.map((key) => (
                <Bar key={key} dataKey={key} fill={chartConfig[key].color} />
              ))}
              <Brush
                dataKey="period"
                height={10}
                stroke="var(--color-accent)"
                fill="var(--color-background)"
              />
              <ChartLegend content={<ChartLegendContent />} />
            </BarChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
