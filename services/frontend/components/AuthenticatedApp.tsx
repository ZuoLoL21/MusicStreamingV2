'use client';

import { Sidebar } from '@/components/Sidebar';
import { Player } from '@/components/Player';

export function AuthenticatedApp({ children }: { children: React.ReactNode }) {
  return (
    <>
      <div className="flex h-screen bg-black text-white">
        <Sidebar />
        <main className="flex-1 overflow-y-auto pb-24">
          {children}
        </main>
      </div>
      <Player />
    </>
  );
}
