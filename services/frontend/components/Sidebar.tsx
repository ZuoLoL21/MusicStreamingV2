'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Cookies from 'js-cookie';
import { useAuthStore } from '@/lib/store';
import {
  Home,
  Search,
  Library,
  PlusSquare,
  Heart,
  Music,
  Compass,
  Clock,
  User,
  LogIn,
  LogOut,
} from 'lucide-react';

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const { clearAuth } = useAuthStore();

  useEffect(() => {
    // Check if user has a token
    const token = Cookies.get('token');
    setIsLoggedIn(!!token);
  }, [pathname]); // Re-check on route changes

  const publicNavItems = [
    { href: '/', label: 'Home', icon: Home },
    { href: '/search', label: 'Search', icon: Search },
    { href: '/discover', label: 'Discover', icon: Compass },
  ];

  const authenticatedNavItems = [
    { href: '/library', label: 'Your Library', icon: Library },
  ];

  const handleLogout = () => {
    // Clear cookies with proper path
    Cookies.remove('token', { path: '/' });
    Cookies.remove('refresh_token', { path: '/' });
    Cookies.remove('user_uuid', { path: '/' });

    // Clear Zustand store
    clearAuth();

    setIsLoggedIn(false);

    // Redirect to home (which will show landing page)
    router.push('/');
  };

  const isActive = (path: string) => {
    return pathname === path ? 'text-white' : 'text-gray-400 hover:text-white';
  };

  return (
    <aside className="w-64 bg-black p-6 flex flex-col border-r border-gray-800">
      <div className="mb-8">
        <Link href="/" className="flex items-center space-x-2">
          <Music className="w-8 h-8 text-green-500" />
          <span className="text-xl font-bold">MusicStream</span>
        </Link>
      </div>

      <nav className="flex-1">
        {/* Public Navigation */}
        <ul className="space-y-4">
          {publicNavItems.map((item) => {
            const Icon = item.icon;
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={`flex items-center space-x-3 transition-colors ${isActive(item.href)}`}
                >
                  <Icon className="w-6 h-6" />
                  <span className="font-semibold">{item.label}</span>
                </Link>
              </li>
            );
          })}

          {/* Authenticated Navigation */}
          {isLoggedIn && authenticatedNavItems.map((item) => {
            const Icon = item.icon;
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={`flex items-center space-x-3 transition-colors ${isActive(item.href)}`}
                >
                  <Icon className="w-6 h-6" />
                  <span className="font-semibold">{item.label}</span>
                </Link>
              </li>
            );
          })}
        </ul>

        {/* Authenticated Features */}
        {isLoggedIn ? (
          <div className="mt-8 pt-8 border-t border-gray-800">
            <ul className="space-y-4">
              <li>
                <Link
                  href="/profile"
                  className={`flex items-center space-x-3 transition-colors ${isActive('/profile')}`}
                >
                  <User className="w-6 h-6" />
                  <span>Your Profile</span>
                </Link>
              </li>
              <li>
                <Link
                  href="/playlists/create"
                  className="flex items-center space-x-3 text-gray-400 hover:text-white transition-colors"
                >
                  <PlusSquare className="w-6 h-6" />
                  <span>Create Playlist</span>
                </Link>
              </li>
              <li>
                <Link
                  href="/liked"
                  className={`flex items-center space-x-3 transition-colors ${isActive('/liked')}`}
                >
                  <Heart className="w-6 h-6" />
                  <span>Liked Songs</span>
                </Link>
              </li>
              <li>
                <Link
                  href="/profile/history"
                  className={`flex items-center space-x-3 transition-colors ${isActive('/profile/history')}`}
                >
                  <Clock className="w-6 h-6" />
                  <span>History</span>
                </Link>
              </li>
            </ul>
          </div>
        ) : null}
      </nav>

      {/* Login/Logout Button */}
      <div className="mt-auto pt-4 border-t border-gray-800">
        {isLoggedIn ? (
          <button
            onClick={handleLogout}
            className="flex items-center space-x-3 text-gray-400 hover:text-white transition-colors w-full"
          >
            <LogOut className="w-6 h-6" />
            <span>Log Out</span>
          </button>
        ) : (
          <Link
            href="/login"
            className="flex items-center space-x-3 text-gray-400 hover:text-white transition-colors"
          >
            <LogIn className="w-6 h-6" />
            <span>Log In</span>
          </Link>
        )}
      </div>
    </aside>
  );
}
