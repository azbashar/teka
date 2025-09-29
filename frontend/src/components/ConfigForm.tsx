"use client";

import * as React from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";

import { useConfig, useConfigActions } from "@/context/ConfigContext";

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { CircleX, PlusCircle } from "lucide-react";
import { Tooltip, TooltipTrigger, TooltipContent } from "./ui/tooltip";
import { Switch } from "./ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";

const locales = [
  "af-ZA",
  "sq-AL",
  "ar-SA",
  "hy-AM",
  "bn-BD",
  "bn-IN",
  "bs-BA",
  "ca-ES",
  "hr-HR",
  "cs-CZ",
  "da-DK",
  "nl-NL",
  "en-US",
  "eo-001",
  "et-EE",
  "tl-PH",
  "fi-FI",
  "fr-FR",
  "de-DE",
  "el-GR",
  "gu-IN",
  "hi-IN",
  "hu-HU",
  "is-IS",
  "id-ID",
  "it-IT",
  "ja-JP",
  "jw-ID",
  "km-KH",
  "kn-IN",
  "ko-KR",
  "la-VA",
  "lv-LV",
  "lt-LT",
  "mk-MK",
  "ml-IN",
  "mr-IN",
  "my-MM",
  "ne-NP",
  "pl-PL",
  "pt-PT",
  "pa-IN",
  "ro-RO",
  "ru-RU",
  "sr-RS",
  "si-LK",
  "sk-SK",
  "sl-SI",
  "es-ES",
  "su-ID",
  "sw-KE",
  "sv-SE",
  "ta-IN",
  "te-IN",
  "th-TH",
  "tr-TR",
  "uk-UA",
  "vi-VN",
  "cy-GB",
  "xh-ZA",
  "zu-ZA",
];

// --- Zod Schema mirroring Config ---
const configSchema = z.object({
  BaseCurrency: z.string().min(1),
  Locale: z.string().min(1),
  AmountColumn: z.number().min(1),
  Accounts: z.object({
    ConversionAccount: z.string().min(1),
    FXGainAccount: z.string().min(1),
    FXLossAccount: z.string().min(1),
    IncomeAccount: z.string().min(1),
    ExpenseAccount: z.string().min(1),
    AssetsAccount: z.string().min(1),
    LiabilitiesAccount: z.string().min(1),
    EquityAccount: z.string().min(1),
  }),
  EfficientFileStructure: z.object({
    Enabled: z.boolean(),
    FilesRoot: z.string().min(1),
  }),
  StarredAccounts: z
    .array(
      z.object({
        DisplayName: z.string().min(1),
        Account: z.string().min(1),
      })
    )
    .min(1, "At least one starred account is required"),

  ShowGetStarted: z.boolean(),
});

type ConfigFormValues = z.infer<typeof configSchema>;

// --- Component ---
export function ConfigForm() {
  const config = useConfig();

  const { updateConfig } = useConfigActions();

  const form = useForm<ConfigFormValues>({
    resolver: zodResolver(configSchema),
    defaultValues: config ?? undefined, // prepopulate from context
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "StarredAccounts",
  });
  const isEfficientEnabled = form.watch("EfficientFileStructure.Enabled");

  async function onSubmit(values: ConfigFormValues) {
    try {
      values.ShowGetStarted = false;
      await updateConfig(values);
      form.reset(values); // reset form state to reflect saved values
    } catch (err) {
      console.error("Failed to save config", err);
    }
  }

  if (!config) return null; // or show loading spinner

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="space-y-8 max-w-2xl"
      >
        {/* BaseCurrency */}
        <FormField
          control={form.control}
          name="BaseCurrency"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Base Currency</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Locale */}
        <FormField
          control={form.control}
          name="Locale"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Locale</FormLabel>
              <Select onValueChange={field.onChange} value={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a locale" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {locales.map((locale) => (
                    <SelectItem key={locale} value={locale}>
                      {locale}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* AmountColumn */}
        <FormField
          control={form.control}
          name="AmountColumn"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Amount Column</FormLabel>
              <FormControl>
                <Input
                  type="number"
                  {...field}
                  value={field.value ?? ""}
                  onChange={(e) => field.onChange(Number(e.target.value))}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Accounts Section */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Accounts</h2>
          {(
            Object.keys(
              config.Accounts
            ) as (keyof ConfigFormValues["Accounts"])[]
          ).map((accountKey) => (
            <FormField
              key={accountKey}
              control={form.control}
              name={`Accounts.${accountKey}`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    {accountKey
                      .replace(/FX/g, "Foreign Exchange")
                      .replace(/([A-Z])/g, " $1")
                      .replace(/^./, (str) => str.toUpperCase())
                      .trim()}
                  </FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
        </div>

        {/* EfficientFileStructure */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Efficient File Structure</h2>
          <FormField
            control={form.control}
            name="EfficientFileStructure.Enabled"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center gap-2">
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
                <FormLabel>Enable</FormLabel>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="EfficientFileStructure.FilesRoot"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Files Root</FormLabel>
                <FormControl>
                  <Input {...field} disabled={!isEfficientEnabled} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        {/* Starred Accounts */}
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">Starred Accounts</h2>
          {fields.map((item, index) => (
            <div key={item.id} className="flex gap-4 items-end">
              <FormField
                control={form.control}
                name={`StarredAccounts.${index}.DisplayName`}
                render={({ field }) => (
                  <FormItem className="flex-1">
                    <FormLabel>Display Name</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name={`StarredAccounts.${index}.Account`}
                render={({ field }) => (
                  <FormItem className="flex-1">
                    <FormLabel>Account</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => remove(index)}
                    aria-label="Remove Starred Account"
                  >
                    <CircleX />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Remove this starred account</TooltipContent>
              </Tooltip>
            </div>
          ))}

          <Button
            type="button"
            variant="ghost"
            onClick={() => append({ DisplayName: "", Account: "" })}
            className="w-full"
          >
            <PlusCircle />
            Add Starred Account
          </Button>
        </div>

        {/* ShowGetStarted */}
        <div className="hidden">
          <FormField
            control={form.control}
            name="ShowGetStarted"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center gap-2">
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
                <FormLabel>Show Get Started Page on Launch</FormLabel>
              </FormItem>
            )}
          />
        </div>

        {/* Submit */}
        <Button className="w-full" type="submit">
          Save Config
        </Button>
      </form>
    </Form>
  );
}
