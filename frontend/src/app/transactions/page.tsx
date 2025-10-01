"use client";
import { DateRange } from "react-day-picker";
import { DateRangePicker } from "@/components/DateRangePicker";
import { usePageTitle } from "@/context/PageTitleContext";
import React from "react";
import TransactionList from "@/components/TransactionList";
import { Input } from "@/components/ui/input";
import { SearchIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

export default function TransactionsPage() {
  const { setTitle } = usePageTitle();
  React.useEffect(() => {
    setTitle("Configuration");
  }, [setTitle]);

  const [range, setRange] = React.useState<DateRange | undefined>({
    from: new Date(new Date().setMonth(new Date().getMonth() - 1)),
    to: new Date(),
  });

  const [accountFeild, setAccountFeild] = React.useState("");
  const [account, setAccount] = React.useState("");

  return (
    <>
      <div className="mb-4">
        <div className="flex flex-col sm:flex-row-reverse justify-between gap-4 sm:items-center mb-2">
          <DateRangePicker range={range} onChange={setRange} />
          <div className="flex items-baseline gap-1">
            <h2 className="font-semibold text-xl">Transactions</h2>
          </div>
        </div>
        <Card>
          <CardHeader className="px-2 sm:px-6">
            <div className="flex mb-4 gap-4">
              <Input
                type="text"
                placeholder="Search..."
                className="flex-1"
                value={accountFeild}
                onChange={(e) => setAccountFeild(e.target.value)}
              />
              <Button
                variant="outline"
                onClick={() => setAccount(accountFeild)}
              >
                <SearchIcon />
                Search
              </Button>
            </div>
          </CardHeader>
          <CardContent className="px-2 sm:px-6">
            <TransactionList range={range} account={account} />
          </CardContent>
        </Card>
      </div>
    </>
  );
}
