"use client";

import { ANSILogo } from "@/components/ANSILogo";
import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { toast } from "sonner"; // using shadcn's toast lib

// --- Types ---
export type Config = {
  BaseCurrency: string;
  Locale: string;
  AmountColumn: number;
  Accounts: {
    ConversionAccount: string;
    FXGainAccount: string;
    FXLossAccount: string;
    IncomeAccount: string;
    ExpenseAccount: string;
    AssetsAccount: string;
    LiabilitiesAccount: string;
    EquityAccount: string;
  };
  EfficientFileStructure: {
    Enabled: boolean;
    FilesRoot: string;
  };
  StarredAccounts: {
    DisplayName: string;
    Account: string;
  }[];
  ShowGetStarted: boolean;
};

// --- Contexts ---
const ConfigContext = createContext<Config | null>(null);
const ConfigActionsContext = createContext<{
  updateConfig: (cfg: Config) => Promise<void>;
} | null>(null);

// --- Provider ---
export const ConfigProvider = ({ children }: { children: ReactNode }) => {
  const [config, setConfig] = useState<Config | null>(null);
  const [loading, setLoading] = useState(true);
  const [status, setStatus] = useState("Loading...");

  useEffect(() => {
    fetch("http://localhost:8080/api/getConfig/")
      .then((res) => res.json())
      .then((data) => {
        setConfig(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Error fetching config:", err);
        setStatus(`Error fetching config.\n ${err}`);
      });
  }, []);

  async function updateConfig(cfg: Config) {
    fetch("http://localhost:8080/api/updateConfig/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(cfg),
    })
      .then((res) => {
        if (!res.ok) {
          res.text().then((err) => {
            throw new Error(`${res.status} ${res.statusText}: ${err}`);
          });
        }
        return res.json();
      })
      .then((data) => {
        setConfig(data);
        console.log("Config updated:", data);
        toast.success("Configuration saved.");
      })
      .catch((err) => {
        console.error("Error updating config:", err);
        toast.error(`Error updating config:\n ${err}`);
      });
  }

  if (loading) {
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-8">
        <ANSILogo className="text-chart-2 max-w-[400px] px-16 min-w-80" />

        <p className="whitespace-pre-line text-center px-4">{status}</p>
      </div>
    );
  }

  return (
    <ConfigContext.Provider value={config}>
      <ConfigActionsContext.Provider value={{ updateConfig }}>
        {children}
      </ConfigActionsContext.Provider>
    </ConfigContext.Provider>
  );
};

// --- Hooks ---
export const useConfig = () => useContext(ConfigContext);

// New hook for updating config
export const useConfigActions = () => {
  const ctx = useContext(ConfigActionsContext);
  if (!ctx) {
    throw new Error("useConfigActions must be used within ConfigProvider");
  }
  return ctx;
};
