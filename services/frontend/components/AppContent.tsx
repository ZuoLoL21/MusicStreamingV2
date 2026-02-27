'use client';

import { useAuth } from '@/components/AuthProvider';
import { LandingPage } from '@/components/LandingPage';
import { AuthenticatedApp } from '@/components/AuthenticatedApp';
import { usePathname } from 'next/navigation';

export function AppContent({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const pathname = usePathname();

  // Show loading spinner while checking auth
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-black">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-green-500"></div>
      </div>
    );
  }

  // If on login page, always show the page content (not the landing page)
  if (pathname === '/login') {
    return <>{children}</>;
  }

  // If not authenticated, show landing page
  if (!isAuthenticated) {
    return <LandingPage />;
  }

  // If authenticated, show the full app with sidebar and player
  return <AuthenticatedApp>{children}</AuthenticatedApp>;
}
