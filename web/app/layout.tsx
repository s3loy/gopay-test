import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'GoPay - Payment System',
  description: 'Full-stack payment system with WeChat and Alipay',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="zh-CN">
      <body className="min-h-screen bg-gray-50">
        <nav className="bg-white shadow-sm border-b">
          <div className="max-w-6xl mx-auto px-4 py-4">
            <h1 className="text-xl font-bold text-gray-900">GoPay 收银台</h1>
          </div>
        </nav>
        <main className="max-w-6xl mx-auto px-4 py-8">{children}</main>
      </body>
    </html>
  );
}
