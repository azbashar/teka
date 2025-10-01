"use client";
import { Plus_Jakarta_Sans, Lora, IBM_Plex_Mono } from "next/font/google";
import "@/app/globals.css";
import { ThemeProvider } from "@/components/ThemeProvider";
import { ConfigProvider, useConfig } from "@/context/ConfigContext";
import { Toaster } from "@/components/ui/sonner";
import { AppSidebar } from "@/components/AppSidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { PageTitleProvider } from "@/context/PageTitleContext";
import { SiteHeader } from "@/components/SiteHeader";
import { ConfigForm } from "@/components/ConfigForm";
import { ANSILogo } from "@/components/ANSILogo";

const plusJakartaSans = Plus_Jakarta_Sans({
  variable: "--font-sans",
  subsets: ["latin"],
});

const lora = Lora({
  variable: "--font-serif",
  subsets: ["latin"],
});

const ibmPlexMono = IBM_Plex_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
  weight: "400",
});

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <title>Teka Finance</title>
      </head>
      <body
        className={`${plusJakartaSans.variable} ${lora.variable} ${ibmPlexMono.variable} antialiased`}
      >
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <ConfigProvider>
            <ConfigConsumer>{children}</ConfigConsumer>
          </ConfigProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}

const ConfigConsumer = ({ children }: { children: React.ReactNode }) => {
  const config = useConfig();
  return (
    <>
      {config?.ShowGetStarted ? (
        <div className="min-h-screen min-w-screen flex flex-col justify-center items-center">
          <Toaster />
          <main className="w-full h-full">
            <div className="w-full h-full flex justify-center items-center py-16">
              <div className="w-full max-w-[500px] flex flex-col justify-center items-center gap-8">
                <ANSILogo className="text-chart-1 max-w-[400px] px-16 min-w-80" />
                <div className="flex flex-col gap-2 max-w-[400px]">
                  <p className="text-center">Welcome to Teka!</p>
                  <p className="text-center text-muted-foreground">
                    Before you start using Teka, you need to configure it so
                    that Teka can properly read your journals.
                  </p>
                </div>
                <div className="border rounded-md p-4 bg-card mt-4 w-full">
                  <h1 className="text-center text-xl font-semibold mb-4">
                    Configuration
                  </h1>
                  <ConfigForm />
                </div>
              </div>
            </div>
          </main>
        </div>
      ) : (
        <SidebarProvider>
          <Toaster />
          <AppSidebar />
          <SidebarInset>
            <PageTitleProvider>
              <SiteHeader />
              <main className="px-4 py-2">{children}</main>
            </PageTitleProvider>
          </SidebarInset>
        </SidebarProvider>
      )}
    </>
  );
};
