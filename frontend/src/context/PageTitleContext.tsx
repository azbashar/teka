"use client";
import { createContext, useContext, useState } from "react";

const PageTitleContext = createContext({
  title: "Teka Finance",
  setTitle: (t: string) => {},
});

export function PageTitleProvider({ children }: { children: React.ReactNode }) {
  const [title, setTitle] = useState("Teka Finance");
  return (
    <PageTitleContext.Provider value={{ title, setTitle }}>
      {children}
    </PageTitleContext.Provider>
  );
}

export const usePageTitle = () => useContext(PageTitleContext);
