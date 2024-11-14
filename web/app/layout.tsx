import type { Metadata } from "next";
import "./globals.css";
import { NextUIProvider } from "@nextui-org/system";

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
