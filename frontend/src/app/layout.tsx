import type { Metadata } from "next";
import { Akatab, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/AppSidebar";
import { SiteHeader } from "@/components/SiteHeader";
import { ThemeProvider } from "@/components/ThemeProvider";
import { PageTitleProvider } from "@/context/PageTitleContext";
import { ConfigProvider } from "@/context/ConfigContext";

const jetBrainsMono = JetBrains_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
});

const akatab = Akatab({
  weight: "600",
  variable: "--font-sans",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Teka Finance",
  description: "Visualize your personal finances",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${akatab.variable} ${jetBrainsMono.variable} antialiased`}
      >
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <ConfigProvider>
            <SidebarProvider>
              <AppSidebar />
              <SidebarInset>
                <PageTitleProvider>
                  <SiteHeader />
                  <main className="px-4 py-2">{children}</main>
                </PageTitleProvider>
              </SidebarInset>
            </SidebarProvider>
          </ConfigProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
