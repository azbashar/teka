"use client";
import { createContext, useContext, useEffect, useState } from "react";

const ConfigContext = createContext<Config | null>(null);

import { ReactNode } from "react";

export type Config = {
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
};

export const ConfigProvider = ({ children }: { children: ReactNode }) => {
  const [config, setConfig] = useState<Config | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("http://localhost:8080/api/getConfig/")
      .then((res) => res.json())
      .then((data) => setConfig(data))
      .catch((err) => console.error("Failed to fetch config:", err))
      .finally(() => setLoading(false));
  }, []);
  if (loading) {
    // prettier-ignore
    const ascii = `████████╗███████╗██╗  ██╗ █████╗     
╚══██╔══╝██╔════╝██║ ██╔╝██╔══██╗    
   ██║   █████╗  █████╔╝ ███████║    
   ██║   ██╔══╝  ██╔═██╗ ██╔══██║    
   ██║   ███████╗██║  ██╗██║  ██║    
   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝    `;
    return (
      <div className="w-screen h-screen flex flex-col items-center justify-center gap-4">
        <pre
          style={{
            whiteSpace: "pre",
            fontFamily: "ui-monospace, Menlo, monospace",
            lineHeight: 1,
          }}
          className="text-chart-2"
        >
          {ascii}
        </pre>

        <p>Loading...</p>
      </div>
    );
  }
  return (
    <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
  );
};

export const useConfig = () => {
  return useContext(ConfigContext);
};
