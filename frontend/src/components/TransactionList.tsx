"use client";

import React from "react";
import { DateRange } from "react-day-picker";
import { Card, CardContent } from "./ui/card";
import { Separator } from "./ui/separator";
import { formatLocalDate } from "@/lib/utils";
import { useConfig } from "@/context/ConfigContext";
import { Tooltip, TooltipContent, TooltipTrigger } from "./ui/tooltip";
import {
  ArrowLeftRight,
  AtSignIcon,
  Calendar,
  CircleAlert,
  CircleCheck,
  Hash,
  HashIcon,
  Info,
  PaperclipIcon,
  TableProperties,
  Tag,
} from "lucide-react";
import { Badge } from "./ui/badge";
import { toast } from "sonner";
import { Button } from "./ui/button";

type TransactionListProps = {
  range: DateRange | undefined;
  account?: string;
  valueMode?: "then" | "end";
};

type Transaction = {
  id: number;
  date: string;
  description: string;
  tags: { key: string; value: string }[];
  comment: string;
  code: string;
  status: string;
  doc: { attached: boolean; path: string };
  postings: {
    account: string;
    amount: number;
    commodity: string;
    cost: { hasCost: boolean; amount: number; commodity: string };
    comment: string;
    status: string;
    tags: { key: string; value: string }[];
  }[];
};

export default function TransactionList({
  range,
  account,
  valueMode,
}: TransactionListProps) {
  const [transactions, setTransactions] = React.useState<Transaction[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [visibleCount, setVisibleCount] = React.useState(15);

  React.useEffect(() => {
    setLoading(true);

    const startDate = formatLocalDate(range?.from);
    const endDate = formatLocalDate(range?.to);
    const acct = account || "";
    const value = valueMode || "";
    fetch(
      `http://localhost:8080/api/transactions/?startDate=${startDate}&endDate=${endDate}&valueMode=${value}&account=${acct}`
    )
      .then((res) => {
        if (!res.ok) {
          return res.text().then((body) => {
            throw new Error(`(${res.status}) ${res.statusText} : ${body}`);
          });
        }
        return res.json();
      })
      .then((data) => {
        setTransactions(data.transactions);
        setLoading(false);
      })
      .catch((err) => {
        toast.error(`Error fetching data: ${err.message}`);
        console.error(err);
      });
  }, [range, account, valueMode]);

  return (
    <div className="flex flex-col items-center justify-center ">
      {!loading ? (
        !transactions || transactions.length < 1 ? (
          <p className="text-center text-muted-foreground">
            No transactions during selected date range.
          </p>
        ) : (
          <div className="w-full">
            <div className="hidden lg:flex gap-4 pb-2 px-4">
              <div className="flex gap-2 w-64">
                <Calendar className="w-4" />
                Date
              </div>
              <div className="w-full flex gap-2">
                <ArrowLeftRight className="w-4" />
                Transaction
              </div>
              <div className="w-full flex gap-2">
                <TableProperties className="w-4" />
                Postings
              </div>
            </div>
            <div className=" mb-2 hidden lg:block">
              <Separator />
            </div>
            {transactions.slice(0, visibleCount).map((tx, i) => {
              return (
                <div className="w-full" key={i}>
                  <TransactionItem transaction={tx} />
                </div>
              );
            })}
            {visibleCount < transactions.length && (
              <div className="w-full flex justify-center">
                <Button
                  variant="outline"
                  onClick={() => setVisibleCount(visibleCount + 15)}
                >
                  Load more
                </Button>
              </div>
            )}
          </div>
        )
      ) : (
        <p className="text-center text-muted-foreground">Loading...</p>
      )}
    </div>
  );
}

function TransactionItem({ transaction }: { transaction: Transaction }) {
  const config = useConfig();
  transaction.postings.forEach((posting) => {
    posting.tags = posting.tags.filter(
      (tag) => tag.key !== "doc" && tag.key !== "type"
    );
  });

  return (
    <div className="border rounded-md px-4 py-2 mb-3">
      <div className="flex flex-col lg:flex-row gap-2 lg:gap-4 py-2">
        <div className="w-64 text-muted-foreground text-sm">
          {new Date(transaction.date).toLocaleDateString(config?.Locale, {
            year: "numeric",
            month: "short",
            day: "numeric",
          })}
        </div>
        <div className="w-full">
          <div>
            {transaction.status.toLowerCase() == "pending" && (
              <span className="inline-flex mr-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <CircleAlert className="size-3 text-yellow-300" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Pending</p>
                  </TooltipContent>
                </Tooltip>
              </span>
            )}
            {transaction.status.toLowerCase() == "cleared" && (
              <span className="inline-flex mr-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <CircleCheck className="size-3 text-green-300" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Cleared</p>
                  </TooltipContent>
                </Tooltip>
              </span>
            )}
            <span>{transaction.description}</span>
            {transaction.code && (
              <span className="inline-flex ml-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Hash className="size-3" />
                  </TooltipTrigger>
                  <TooltipContent>
                    Code: <code>{transaction.code}</code>
                  </TooltipContent>
                </Tooltip>
              </span>
            )}
            {transaction.doc.attached && (
              <span className="inline-flex ml-1">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <PaperclipIcon className="size-3" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{transaction.doc.path}</p>
                  </TooltipContent>
                </Tooltip>
              </span>
            )}
          </div>
          {stripHledgerTags(transaction.comment) && (
            <div className="text-sm text-muted-foreground">
              <CollapsibleComment
                comment={stripHledgerTags(transaction.comment)}
              />
            </div>
          )}
          {transaction.tags && transaction.tags.length > 0 && (
            <div className="flex gap-2 flex-wrap text-sm mt-2">
              {transaction.tags.map((tag, i) => {
                if (tag.key != "doc") {
                  return (
                    <div key={i}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Badge variant="secondary" className="flex gap-1">
                            <Tag className="size-3" />
                            {tag.key}
                          </Badge>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>
                            {tag.key}: {tag.value}
                          </p>
                        </TooltipContent>
                      </Tooltip>
                    </div>
                  );
                }
              })}
            </div>
          )}
        </div>
        <div className="lg:hidden">
          <Separator className="lg:hidden" />
        </div>
        <div className="w-full">
          <table className="w-full table-fixed text-sm">
            <tbody>
              {transaction.postings.map((posting, i) => {
                return (
                  <tr className={i % 2 === 0 ? "bg-muted" : ""} key={i}>
                    <td>
                      <div className="overflow-auto w-full pr-2 ">
                        {posting.status.toLowerCase() == "pending" && (
                          <span className="inline-flex mr-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <CircleAlert className="size-3 text-yellow-300" />
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Pending</p>
                              </TooltipContent>
                            </Tooltip>
                          </span>
                        )}
                        {posting.status.toLowerCase() == "cleared" && (
                          <span className="inline-flex mr-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <CircleCheck className="size-3 text-green-300" />
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>Cleared</p>
                              </TooltipContent>
                            </Tooltip>
                          </span>
                        )}
                        <span>{posting.account}</span>

                        {stripHledgerTags(posting.comment) && (
                          <span className="inline-flex ml-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Info className="size-3" />
                              </TooltipTrigger>
                              <TooltipContent>
                                <CollapsibleComment
                                  comment={stripHledgerTags(posting.comment)}
                                />
                              </TooltipContent>
                            </Tooltip>
                          </span>
                        )}
                        {posting.tags && posting.tags.length > 0 && (
                          <span className="inline-flex ml-1">
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Tag className="size-3" />
                              </TooltipTrigger>
                              <TooltipContent>
                                {posting.tags.map((tag, i) => {
                                  return (
                                    <p key={i}>
                                      {tag.key}: {tag.value}
                                    </p>
                                  );
                                })}
                              </TooltipContent>
                            </Tooltip>
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="w-1/3 text-right whitespace-nowrap">
                      <div className="overflow-x-scroll w-full font-mono flex gap-1 justify-end items-center">
                        <p className="w-full">
                          {posting.amount.toLocaleString(config?.Locale)}{" "}
                          {posting.commodity}
                        </p>
                        {posting.cost.hasCost && (
                          <span>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <AtSignIcon className="size-3" />
                              </TooltipTrigger>
                              <TooltipContent>
                                <p>
                                  Cost:{" "}
                                  {posting.cost.amount.toLocaleString(
                                    config?.Locale
                                  )}{" "}
                                  {posting.cost.commodity}
                                </p>
                              </TooltipContent>
                            </Tooltip>
                          </span>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function CollapsibleComment({ comment }: { comment: string }) {
  const [open, setOpen] = React.useState(false);
  const toggleOpen = () => setOpen(!open);
  if (comment.length <= 70) return <p>{comment}</p>;
  return (
    <>
      {open ? (
        <>
          <p>{comment}</p>
          <p className="underline cursor-pointer" onClick={toggleOpen}>
            See less
          </p>
        </>
      ) : (
        <>
          <p>{comment.slice(0, 70)}...</p>
          <p className="underline cursor-pointer" onClick={toggleOpen}>
            See more
          </p>
        </>
      )}
    </>
  );
}

function stripHledgerTags(comment: string) {
  if (!comment) return comment;

  return (
    comment
      // Remove tags with optional values, including leading spaces or commas
      .replace(
        /(?:^|[\s,])([^\s,]+:)([^,\n]*)/g,
        (match, tag, value, offset) => {
          // If tag is at start of string, just remove it
          if (offset === 0) return "";
          // Otherwise remove the match but keep preceding space if it exists
          return match[0] === " " ? " " : "";
        }
      )
      // Remove leftover commas and extra spaces from removed tags
      .replace(/,+/g, " ")
      .replace(/\s+/g, " ")
      .trim()
  );
}
