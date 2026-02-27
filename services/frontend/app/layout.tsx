import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from '@/components/AuthProvider';
import { AppContent } from '@/components/AppContent';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'MusicStream - Your Music Platform',
  description: 'Stream your favorite music',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <AuthProvider>
          <AppContent>{children}</AppContent>
        </AuthProvider>
        <Toaster position="top-right" />
      </body>
    </html>
  );
}