import type { Metadata } from "next";
import "./globals.css";
import { TimeSyncedBackground } from "@/components/TimeSyncedBackground";
import { Header } from "@/components/Header";
import { RightSidebar } from "@/components/RightSidebar";
import { CSPostHogProvider } from "@/providers/PostHogProvider";
import { AuthProvider } from "@/providers/AuthProvider";
import { QueryProvider } from "@/providers/QueryProvider";

export const metadata: Metadata = {
  title: "Aurea | AI Commerce Ops",
  description: "Next Generation AI Commerce Dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className="h-full antialiased transition-colors duration-1000"
    >
      <body className="min-h-screen flex items-center justify-center p-3 sm:p-4 lg:p-5 bg-transparent overflow-hidden text-foreground">
        <CSPostHogProvider>
          <AuthProvider>
            <QueryProvider>
              <TimeSyncedBackground />
              
              <div className="relative w-full h-[96vh] flex flex-col z-10">
                
                {/* The main frosted glass backdrop that fades out on the right */}
                <div className="absolute inset-0 bg-white/85 dark:bg-[#0a0a0a]/85 backdrop-blur-[50px] rounded-[2.5rem] -z-10 shadow-2xl ring-1 ring-black/5 dark:ring-white/10" style={{
                  maskImage: 'linear-gradient(to right, black 65%, transparent 100%)',
                  WebkitMaskImage: 'linear-gradient(to right, black 65%, transparent 100%)'
                }} />
                
                {/* Header spans the entire top width */}
                <Header />
                
                {/* Below Header: Main Content Area */}
                <div className="flex-1 flex overflow-hidden px-10 pb-10 gap-10">
                  {children}
                </div>
                
              </div>
            </QueryProvider>
          </AuthProvider>
        </CSPostHogProvider>
      </body>
    </html>
  );
}

