"use client";
import { BalanceCard } from "@/components/BalanceCard";
import { DateRangePicker } from "@/components/DateRangePicker";
import { ExpensePieChart } from "@/components/ExpensePieChart";
import { IncomePieChart } from "@/components/IncomePieChart";
import { NetWorthChart } from "@/components/NetWorthChart";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { HelpCircle } from "lucide-react";
import * as React from "react";
import { DateRange } from "react-day-picker";

// get account balances from api and show them as card
async function getAccountBalances() {
  const res = await fetch("http://localhost:8080/api/accountBalances/");
  const data = await res.json();
  return data;
}

export default function Home() {
  const [range, setRange] = React.useState<DateRange | undefined>({
    from: new Date(new Date().setMonth(new Date().getMonth() - 1)),
    to: new Date(),
  });

  const [accountBalances, setAccountBalances] = React.useState<
    {
      id: string;
      displayName: string;
      account: string;
      balance: string;
      percentChange: string;
    }[]
  >([]);

  React.useEffect(() => {
    const fetchData = async () => {
      const data = await getAccountBalances();
      setAccountBalances(data);
    };
    fetchData();
  }, [range]);

  return (
    <div>
      <title>Dashboard</title>
      <div className="mb-4">
        <div className="flex flex-col sm:flex-row-reverse justify-between gap-4 sm:items-center mb-2">
          <DateRangePicker range={range} onChange={setRange} />
          <div className="flex items-baseline gap-1">
            <h2 className="font-semibold text-xl">Account Balances</h2>
            <Tooltip>
              <TooltipTrigger>
                <HelpCircle className="size-3 text-gray-200" />
              </TooltipTrigger>
              <TooltipContent>
                <p>
                  Accounts added to <code>starred_accounts</code> in the config
                  will be shown here.
                </p>
              </TooltipContent>
            </Tooltip>
          </div>
        </div>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
          {accountBalances.map(
            (account: {
              id: string;
              displayName: string;
              account: string;
              balance: string;
              percentChange: string;
            }) => (
              <BalanceCard
                accountName={account.displayName}
                account={account.account}
                balance={account.balance}
                percentChange={account.percentChange}
                key={account.id}
              />
            )
          )}
        </div>
      </div>
      <div className="flex flex-col gap-4">
        <NetWorthChart range={range} />
        <div className="grid lg:grid-cols-2 gap-4">
          <IncomePieChart range={range} />
          <ExpensePieChart range={range} />
        </div>
      </div>
    </div>
  );
}
