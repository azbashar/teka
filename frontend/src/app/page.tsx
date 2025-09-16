import { BalanceCard } from "@/components/BalanceCard";
import { ExpensePieChart } from "@/components/ExpensePieChart";
import { IncomePieChart } from "@/components/IncomePieChart";
import { NetWorthChart } from "@/components/NetWorthChart";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { HelpCircle } from "lucide-react";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Dashboard",
};

// get account balances from api and show them as card
async function getAccountBalances() {
  const res = await fetch("http://localhost:8080/api/accountBalances");
  const data = await res.json();
  return data;
}

export default async function Home() {
  const accountBalances = await getAccountBalances();
  return (
    <div>
      <div className="mb-4">
        <div className="flex items-baseline gap-1 mb-4">
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
        <NetWorthChart />
        <div className="grid lg:grid-cols-2 gap-4">
          <IncomePieChart />
          <ExpensePieChart />
        </div>
      </div>
    </div>
  );
}
