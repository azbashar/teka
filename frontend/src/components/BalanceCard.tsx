import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import { Badge } from "@/components/ui/badge";
import { MinusIcon, TrendingDownIcon, TrendingUpIcon } from "lucide-react";

interface BalanceCardProps {
  accountName: string;
  balance: string;
  account: string;
  percentChange: string;
}

export function BalanceCard({
  accountName,
  balance,
  account,
  percentChange,
}: BalanceCardProps) {
  const change = parseFloat(percentChange).toFixed(0);
  return (
    <Card>
      <CardHeader className="relative">
        <CardDescription>{accountName}</CardDescription>
        <CardTitle className="@[250px]/card:text-3xl text-2xl font-semibold tabular-nums">
          {balance}
        </CardTitle>
        <div className="absolute right-4">
          {/* if change posetive write posetive and if negative write negative and if 0 change then write grey*/}
          {parseFloat(change) > 0 ? (
            <Badge
              variant="outline"
              className="flex gap-1 rounded-lg text-xs text-green-400 opacity-70"
            >
              <TrendingUpIcon className="size-3" />
              {change}%
            </Badge>
          ) : parseFloat(change) < 0 ? (
            <Badge
              variant="outline"
              className="flex gap-1 rounded-lg text-xs text-red-400 opacity-70"
            >
              <TrendingDownIcon className="size-3" />
              {change}%
            </Badge>
          ) : (
            <Badge variant="outline" className="text-gray-500">
              <MinusIcon className="size-3" />
              {change}%
            </Badge>
          )}
        </div>
      </CardHeader>
      <CardFooter className="flex-col items-start gap-1 text-sm">
        <div className="text-muted-foreground">{account}</div>
      </CardFooter>
    </Card>
  );
}
