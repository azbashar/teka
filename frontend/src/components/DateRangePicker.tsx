"use client";

import * as React from "react";
import { DateRange } from "react-day-picker";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Calendar } from "@/components/ui/calendar";
import { CalendarIcon } from "lucide-react";

type DateRangePickerProps = {
  range: DateRange | undefined;
  onChange: (range: DateRange | undefined) => void;
};

export function DateRangePicker({ range, onChange }: DateRangePickerProps) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline">
          <CalendarIcon className="mr-2" />
          {range?.from && range?.to
            ? `${range.from.toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                year: "numeric",
              })} - ${range.to.toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                year: "numeric",
              })}`
            : "Select range"}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[300px] p-0" align="end">
        <Calendar
          className="w-full"
          mode="range"
          selected={range}
          onSelect={onChange}
          captionLayout="dropdown"
        />
        <div className="grid grid-cols-2 gap-2 p-2">
          <Button
            onClick={() =>
              onChange({
                from: new Date(new Date().setDate(new Date().getDate() - 7)),
                to: new Date(),
              })
            }
            className="w-full"
            variant="outline"
          >
            Last 7 days
          </Button>
          <Button
            onClick={() =>
              onChange({
                from: new Date(new Date().setDate(new Date().getDate() - 30)),
                to: new Date(),
              })
            }
            className="w-full"
            variant="outline"
          >
            Last 30 days
          </Button>
          <Button
            onClick={() =>
              onChange({
                from: new Date(new Date().setDate(new Date().getDate() - 90)),
                to: new Date(),
              })
            }
            className="w-full"
            variant="outline"
          >
            Last 90 days
          </Button>
          <Button
            onClick={() =>
              onChange({
                from: new Date(new Date().setDate(new Date().getDate() - 365)),
                to: new Date(),
              })
            }
            className="w-full"
            variant="outline"
          >
            Last 365 days
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
