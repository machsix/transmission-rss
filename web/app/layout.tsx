import { NextUIProvider } from "@nextui-org/react";
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
        <NextUIProvider>
          {children}
        </NextUIProvider>
      </body>
    </html>
  );
}
