"use client";
import { createContext, useContext, useEffect, useState } from "react";

const ConfigContext = createContext<Config | null>(null);

import { ReactNode } from "react";

type Config = {
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

  useEffect(() => {
    fetch("http://localhost:8080/api/getConfig/")
      .then((res) => res.json())
      .then((data) => setConfig(data))
      .catch((err) => console.error("Failed to fetch config:", err));
  }, []);

  return (
    <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
  );
};

export const useConfig = () => {
  return useContext(ConfigContext);
};
