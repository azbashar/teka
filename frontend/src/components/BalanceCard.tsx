import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

interface BalanceCardProps {
  accountName: string;
  balance: string;
  account: string;
}

export function BalanceCard({
  accountName,
  balance,
  account,
}: BalanceCardProps) {
  return (
    <Card className="max-w-64 w-full">
      <CardHeader>
        <CardTitle>{accountName}</CardTitle>
        <CardDescription>{account}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-xl">{balance}</p>
      </CardContent>
    </Card>
  );
}
