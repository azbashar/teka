"use client";
import { createContext, useContext, useEffect, useState } from "react";

const ConfigContext = createContext(null);

import { ReactNode } from "react";

export const ConfigProvider = ({ children }: { children: ReactNode }) => {
  const [config, setConfig] = useState(null);

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
