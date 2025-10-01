"use client";
import { BalanceCard } from "@/components/BalanceCard";
import { DateRangePicker } from "@/components/DateRangePicker";
import { StatementPieChart } from "@/components/StatementPieChart";
import { NetWorthChart } from "@/components/NetWorthChart";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useConfig } from "@/context/ConfigContext";
import { usePageTitle } from "@/context/PageTitleContext";
import { formatLocalDate } from "@/lib/utils";
import { HelpCircle } from "lucide-react";
import * as React from "react";
import { DateRange } from "react-day-picker";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";

// get account balances from api and show them as card
async function getAccountBalances(date: string) {
  const res = await fetch(
    `http://localhost:8080/api/accountBalances/?date=${date}`
  );
  const data = await res.json();
  return data;
}

export default function Home() {
  const config = useConfig();
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
      const data = await getAccountBalances(formatLocalDate(range?.to));
      setAccountBalances(data);
    };
    fetchData();
  }, [range]);

  const { setTitle } = usePageTitle();
  React.useEffect(() => {
    setTitle("Dashboard");
  }, [setTitle]);

  return (
    <div>
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
        <ScrollArea>
          <div className="flex sm:grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
            {accountBalances.map(
              (account: {
                id: string;
                displayName: string;
                account: string;
                balance: string;
                percentChange: string;
              }) => (
                <div className="min-w-[250px] sm:min-w-full" key={account.id}>
                  <BalanceCard
                    accountName={account.displayName}
                    account={account.account}
                    balance={account.balance}
                    percentChange={account.percentChange}
                  />
                </div>
              )
            )}
          </div>
          <ScrollBar orientation="horizontal" />
        </ScrollArea>
      </div>
      <div className="flex flex-col gap-4">
        <NetWorthChart range={range} />
        <div className="grid lg:grid-cols-2 gap-4">
          <StatementPieChart
            range={range}
            title="Income"
            description="Your income distribution."
            rootAccount={config?.Accounts.IncomeAccount}
            statement="incomestatement"
          />
          <StatementPieChart
            range={range}
            title="Expenses"
            description="Your expense distribution."
            rootAccount={config?.Accounts.ExpenseAccount}
            statement="incomestatement"
          />
        </div>
      </div>
    </div>
  );
}
