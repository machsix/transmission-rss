import { HeroUIProvider } from "@heroui/react";
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Transmission RSS",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <HeroUIProvider>
          {children}
        </HeroUIProvider>
      </body>
    </html>
  );
}
