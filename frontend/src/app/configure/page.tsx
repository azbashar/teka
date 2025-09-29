"use client";
import { ConfigForm } from "@/components/ConfigForm";
import { usePageTitle } from "@/context/PageTitleContext";
import React from "react";

export default function ConfigPage() {
  const { setTitle } = usePageTitle();
  React.useEffect(() => {
    setTitle("Configuration");
  }, [setTitle]);
  return (
    <>
      <div className="w-full h-full flex justify-center items-center">
        <div className="w-full max-w-[700px] flex flex-col justify-center items-center gap-8">
          <div className="border rounded p-4 bg-sidebar mt-4 w-full">
            <h1 className="text-center text-xl font-semibold mb-4">
              Configuration
            </h1>
            <ConfigForm />
          </div>
        </div>
      </div>
    </>
  );
}
