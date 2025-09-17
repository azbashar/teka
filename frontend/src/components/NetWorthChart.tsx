"use client";

import * as React from "react";
import { DateRange } from "react-day-picker";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";

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

type NetWorthData = { date: string; networth: number; currency: string };

// function to populate chartData from api request with startDate and endDate params. api request will be localhost:8080/api/networth?startDate=..&endDate=..
async function getNetWorthData(startDate: string, endDate: string) {
  const res = await fetch(
    `http://localhost:8080/api/networth/?startDate=${startDate}&endDate=${endDate}`
  );
  const data = await res.json();
  return data;
}

const chartConfig = {
  visitors: {
    label: "Visitors",
  },
  networth: {
    label: "Networth",
    color: "var(--chart-2)",
  },
} satisfies ChartConfig;

type NetWorthChartProps = {
  range: DateRange | undefined;
};

export function NetWorthChart({ range }: NetWorthChartProps) {
  const [chartData, setChartData] = React.useState<NetWorthData[]>([]);

  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getNetWorthData(
        range?.from?.toISOString().slice(0, 10) || "",
        range?.to?.toISOString().slice(0, 10) || ""
      );
      setChartData(data);
    };
    fetchData();
  }, [range]);

  const filteredData = chartData;

  return (
    <Card className="pt-0">
      <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
        <div className="grid flex-1 gap-1">
          <CardTitle>Net worth</CardTitle>
          <CardDescription>Your total net worth</CardDescription>
        </div>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer
          config={chartConfig}
          className="aspect-auto h-[250px] w-full"
        >
          <AreaChart data={filteredData}>
            <defs>
              <linearGradient id="fillNetworth" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="5%"
                  stopColor="var(--chart-2)"
                  stopOpacity={0.8}
                />
                <stop
                  offset="95%"
                  stopColor="var(--chart-2)"
                  stopOpacity={0.1}
                />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              minTickGap={32}
              tickFormatter={(value) => {
                const date = new Date(value);
                return date.toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                });
              }}
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
              content={
                <ChartTooltipContent
                  labelFormatter={(value) => {
                    return new Date(value).toLocaleDateString("en-US", {
                      month: "short",
                      day: "numeric",
                    });
                  }}
                  formatter={(value) =>
                    `${value.toLocaleString("en-US")} ${
                      chartData[Math.ceil(chartData.length / 2)]?.currency ||
                      chartData[0]?.currency ||
                      ""
                    }`
                  }
                  indicator="dot"
                />
              }
            />
            <Area
              dataKey="networth"
              type="basis"
              fill="url(#fillNetWorth)"
              stroke="var(--chart-2)"
              stackId="a"
            />
            <ChartLegend content={<ChartLegendContent />} />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
