"use client";
import { DateRangePicker } from "@/components/DateRangePicker";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tooltip } from "@/components/ui/tooltip";
import { usePageTitle } from "@/context/PageTitleContext";
import { formatLocalDate } from "@/lib/utils";
import React, { useEffect, useState } from "react";
import { DateRange } from "react-day-picker";
import { Sankey, ResponsiveContainer } from "recharts";
import { toast } from "sonner";

type SankeyNode = { name: string };
type SankeyLink = { source: number; target: number; value: number };
type SankeyData = {
  maxChainLength: number;
  totalSources: number;
  totalTargets: number;
  sankeyData: { nodes: SankeyNode[]; links: SankeyLink[] };
  currency: string;
};

const colors = [
  "var(--color-chart-1)",
  "var(--color-chart-2)",
  "var(--color-chart-3)",
  "var(--color-chart-4)",
  "var(--color-chart-5)",
];

type CustomNodePayload = {
  name: string;
  sourceNodes: number[];
  sourceLinks: number[];
  targetLinks: number[];
  targetNodes: number[];
  value: number;
  depth: number;
  x: number;
  dx: number;
  y: number;
  dy: number;
};

type SankeyNodeProps = {
  x: number;
  y: number;
  width: number;
  height: number;
  payload: CustomNodePayload;
};

const CustomNode = (props: SankeyNodeProps): React.ReactElement<SVGElement> => {
  return (
    <g>
      <rect
        x={props.x + 4}
        y={props.y - 2}
        width={props.width - 8}
        height={props.height + 4}
        fill={colors[props.payload.depth % colors.length]}
        rx={2.5}
      />
      <text
        x={props.x + props.width / 2}
        y={props.y - 10}
        textAnchor="middle"
        alignmentBaseline="middle"
        fill="var(--color-foreground)"
        fontSize={10}
      >
        {props.payload.name}
      </text>
    </g>
  );
};

type CustomLinkPayload = {
  source: CustomNodePayload;
  target: CustomNodePayload;
  value: number;
  dy: number;
  sy: number;
  ty: number;
};

export default function SankeyChart() {
  const { setTitle } = usePageTitle();
  React.useEffect(() => {
    setTitle("Sankey Chart");
  }, [setTitle]);
  const [data, setData] = useState<SankeyData | null>(null);
  const [range, setRange] = React.useState<DateRange | undefined>({
    from: new Date(new Date().setFullYear(new Date().getFullYear() - 1)),
    to: new Date(),
  });
  const [depth, setDepth] = useState("full");

  useEffect(() => {
    fetch(
      `http://localhost:8080/api/sankey/?startDate=${formatLocalDate(
        range?.from
      )}&endDate=${formatLocalDate(
        range?.to
          ? new Date(range.to.getTime() + 24 * 60 * 60 * 1000)
          : new Date()
      )}&depth=${depth == "full" ? "" : depth}`
    )
      .then((res) => {
        if (!res.ok) {
          return res.text().then((body) => {
            throw new Error(`(${res.status}) ${res.statusText} : ${body}}`);
          });
        }
        return res.json();
      })
      .then(setData)
      .catch((err) => {
        toast.error(`Error fetching data: ${err.message}`);
        console.error(
          `Component: SankeyChart, Error fetching data: ${err.message}`
        );
      });
  }, [range, depth]);

  const CustomLink = (props: {
    sourceX: number;
    targetX: number;
    sourceY: number;
    targetY: number;
    sourceControlX: number;
    targetControlX: number;
    sourceRelativeY: number;
    targetRelativeY: number;
    linkWidth: number;
    index: number;
    payload: CustomLinkPayload;
  }) => {
    console.log(props.payload);
    return (
      <g>
        <path
          d={`
  M${props.sourceX},${props.sourceY}
  C${props.sourceControlX},${props.sourceY} ${props.targetControlX},${props.targetY} ${props.targetX},${props.targetY}`}
          fill="none"
          stroke={colors[props.payload.source.depth % colors.length]}
          strokeOpacity={0.4}
          strokeWidth={props.linkWidth}
          strokeLinecap="butt"
        />
        <foreignObject
          x={props.sourceX}
          y={props.targetY - props.linkWidth / 2}
          width={
            Math.max(props.targetX, props.sourceX) -
            Math.min(props.targetX, props.sourceX)
          }
          height={props.linkWidth}
          style={{ overflow: "visible" }}
        >
          <div
            style={{
              boxSizing: "border-box",
              display: "flex",
              alignItems: "center",
              justifyContent: "flex-end",
              width: "100%",
              height: "100%",
              overflow: "visible",
              padding: "0.5em",
              gap: 8,
            }}
          >
            <div
              style={{
                fontSize: 10,
                fontFamily: "sans-serif",
                textAlign: "center",
                padding: "0.25em 0.5em",
                borderRadius: 4,
                position: "relative",
                zIndex: 1,
                opacity: 0.8,
              }}
            >
              {props.payload.value.toLocaleString("en-US", {
                maximumFractionDigits: 2,
                compactDisplay: "short",
                notation: "compact",
              })}{" "}
              {data?.currency ?? ""}
            </div>
          </div>
        </foreignObject>
      </g>
    );
  };

  if (!data)
    return (
      <>
        <div className="flex flex-col sm:flex-row-reverse justify-between gap-4 sm:items-center mb-2">
          <DateRangePicker range={range} onChange={setRange} />
          <div className="flex items-baseline gap-1">
            <h2 className="font-semibold text-xl">Sankey Chart</h2>
          </div>
        </div>
        <div className="flex h-[calc(100vh-8rem))] w-full items-center justify-center  border rounded p-4 bg-sidebar relative">
          <div className="absolute z-50 bg-sidebar rounded-sm top-4 left-4">
            <Select onValueChange={(value) => setDepth(value)} value={depth}>
              <SelectTrigger>
                Depth: <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="full">Full</SelectItem>
                <SelectItem value="1">1</SelectItem>
                <SelectItem value="2">2</SelectItem>
                <SelectItem value="3">3</SelectItem>
                <SelectItem value="4">4</SelectItem>
                <SelectItem value="5">5</SelectItem>
                <SelectItem value="6">6</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </>
    );

  return (
    <>
      <div className="flex flex-col sm:flex-row-reverse justify-between gap-4 sm:items-center mb-2">
        <DateRangePicker range={range} onChange={setRange} />
        <div className="flex items-baseline gap-1">
          <h2 className="font-semibold text-xl">Sankey Chart</h2>
        </div>
      </div>
      <div className="flex w-full">
        <ScrollArea className="w-1 h-[calc(100vh-8rem)] border rounded p-4 bg-sidebar flex-1 relative">
          <div className="absolute z-50 bg-sidebar rounded-sm">
            <Select onValueChange={(value) => setDepth(value)} value={depth}>
              <SelectTrigger>
                Depth: <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="full">Full</SelectItem>
                <SelectItem value="1">1</SelectItem>
                <SelectItem value="2">2</SelectItem>
                <SelectItem value="3">3</SelectItem>
                <SelectItem value="4">4</SelectItem>
                <SelectItem value="5">5</SelectItem>
                <SelectItem value="6">6</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-center items-center min-h-full">
            {data.sankeyData.links.length === 0 && (
              <div className="flex h-[250px] w-full items-center justify-center">
                <p className="text-muted-foreground">
                  No data for selected date range
                </p>
              </div>
            )}
            {data.sankeyData.links.length > 0 &&
              data.sankeyData.nodes.length > 0 && (
                <ResponsiveContainer
                  width={
                    data.maxChainLength * 100 +
                    Math.max(data.totalSources, data.totalTargets) * 15
                  }
                  height={Math.max(
                    data.totalSources * 75,
                    data.totalTargets * 75,
                    500
                  )}
                >
                  <Sankey
                    nameKey="name"
                    data={data.sankeyData}
                    nodePadding={50}
                    margin={{
                      left: 40,
                      right: 60,
                      top: 100,
                      bottom: 100,
                    }}
                    node={CustomNode}
                    link={CustomLink}
                    nodeWidth={15}
                  >
                    <Tooltip />
                  </Sankey>
                </ResponsiveContainer>
              )}
          </div>
          <ScrollBar orientation="horizontal" />
          <ScrollBar orientation="vertical" />
        </ScrollArea>
      </div>
    </>
  );
}
