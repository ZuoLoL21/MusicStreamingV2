'use client';

import { useRouter } from 'next/navigation';
import { Music, Play, Heart, TrendingUp } from 'lucide-react';

export function LandingPage() {
  const router = useRouter();

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-black to-gray-900 flex flex-col">
      {/* Header */}
      <header className="p-6 flex justify-between items-center">
        <div className="flex items-center space-x-2">
          <Music className="w-8 h-8 text-green-500" />
          <span className="text-2xl font-bold text-white">MusicStream</span>
        </div>
        <button
          onClick={() => router.push('/login')}
          className="px-6 py-2 text-white hover:text-green-500 transition-colors font-semibold"
        >
          Log In
        </button>
      </header>

      {/* Hero Section */}
      <main className="flex-1 flex items-center justify-center px-6">
        <div className="max-w-4xl text-center">
          <h1 className="text-6xl md:text-7xl font-bold text-white mb-6 leading-tight">
            Your Music,
            <br />
            <span className="text-green-500">Your Way</span>
          </h1>

          <p className="text-xl md:text-2xl text-gray-300 mb-12 max-w-2xl mx-auto">
            Stream unlimited music, create playlists, discover new artists, and enjoy personalized recommendations.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <button
              onClick={() => router.push('/login')}
              className="px-8 py-4 bg-green-500 hover:bg-green-600 text-white text-lg font-semibold rounded-full transition-all transform hover:scale-105 shadow-lg"
            >
              Get Started
            </button>
            <button
              onClick={() => router.push('/login')}
              className="px-8 py-4 bg-transparent border-2 border-white hover:border-green-500 text-white hover:text-green-500 text-lg font-semibold rounded-full transition-all"
            >
              Sign Up Free
            </button>
          </div>

          {/* Features */}
          <div className="mt-20 grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="flex flex-col items-center text-center">
              <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mb-4">
                <Play className="w-8 h-8 text-green-500" />
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">Stream Anywhere</h3>
              <p className="text-gray-400">
                Listen to your favorite tracks anytime, anywhere
              </p>
            </div>

            <div className="flex flex-col items-center text-center">
              <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mb-4">
                <Heart className="w-8 h-8 text-green-500" />
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">Build Your Library</h3>
              <p className="text-gray-400">
                Create playlists and save your favorite songs
              </p>
            </div>

            <div className="flex flex-col items-center text-center">
              <div className="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mb-4">
                <TrendingUp className="w-8 h-8 text-green-500" />
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">Smart Recommendations</h3>
              <p className="text-gray-400">
                Discover new music tailored to your taste
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="p-6 text-center text-gray-500 text-sm">
        <p>&copy; 2026 MusicStream. Your personal music platform.</p>
      </footer>
    </div>
  );
}
