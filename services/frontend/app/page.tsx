'use client';

import { useEffect, useState } from 'react';
import { api, getFileUrl } from '@/lib/api';
import { formatDuration } from '@/lib/formatters';
import { Music, Artist } from '@/lib/types';
import { Play } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function HomePage() {
  const [recentMusic, setRecentMusic] = useState<Music[]>([]);
  const [artists, setArtists] = useState<Artist[]>([]);
  const [loading, setLoading] = useState(true);
  const { playQueue } = usePlayerStore();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      // Load artists and use their music as "recent"
      const artistsData = await api.getArtists(10);
      setArtists(artistsData);

      // Get music from first artist if available
      if (artistsData.length > 0) {
        const music = await api.getArtistMusic(artistsData[0].uuid, 10);
        setRecentMusic(music);
      }
    } catch (error) {
      toast.error('Failed to load content');
    } finally {
      setLoading(false);
    }
  };

  const handlePlayMusic = (music: Music, index: number) => {
    playQueue(recentMusic, index);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <h1 className="text-4xl font-bold mb-2">Welcome to MusicStream</h1>
      <p className="text-gray-400 mb-8">
        Discover music, create playlists, and share your favorite songs
      </p>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-12">
        <Link
          href="/discover"
          className="bg-gradient-to-br from-purple-600 to-blue-600 p-6 rounded-lg hover:scale-105 transition"
        >
          <h3 className="text-xl font-bold mb-2">Discover</h3>
          <p className="text-sm opacity-90">Explore new music and themes</p>
        </Link>
        <Link
          href="/search"
          className="bg-gradient-to-br from-green-600 to-teal-600 p-6 rounded-lg hover:scale-105 transition"
        >
          <h3 className="text-xl font-bold mb-2">Search</h3>
          <p className="text-sm opacity-90">Find your favorite songs and artists</p>
        </Link>
        <Link
          href="/library"
          className="bg-gradient-to-br from-orange-600 to-red-600 p-6 rounded-lg hover:scale-105 transition"
        >
          <h3 className="text-xl font-bold mb-2">Your Library</h3>
          <p className="text-sm opacity-90">Access your playlists and likes</p>
        </Link>
      </div>

      {/* Artists Section */}
      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">Popular Artists</h2>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
          {artists.map((artist) => (
            <Link
              key={artist.uuid}
              href={`/artists/${artist.uuid}`}
              className="bg-gray-900 p-4 rounded-lg hover:bg-gray-800 transition group"
            >
              <div className="aspect-square bg-gray-800 rounded-full mb-4 flex items-center justify-center overflow-hidden">
                {artist.profile_image_path ? (
                  <img
                    src={getFileUrl(artist.profile_image_path)}
                    alt={artist.artist_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="text-4xl text-gray-600">
                    {artist.artist_name.charAt(0).toUpperCase()}
                  </div>
                )}
              </div>
              <h3 className="font-semibold text-center truncate">{artist.artist_name}</h3>
              <p className="text-sm text-gray-400 text-center">Artist</p>
            </Link>
          ))}
        </div>
      </section>

      {/* Recent Music Section */}
      {recentMusic.length > 0 && (
        <section>
          <h2 className="text-2xl font-bold mb-4">Recent Tracks</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
            {recentMusic.map((music, index) => (
              <div
                key={music.uuid}
                className="bg-gray-900 p-4 rounded-lg hover:bg-gray-800 transition group cursor-pointer"
                onClick={() => handlePlayMusic(music, index)}
              >
                <div className="relative aspect-square bg-gray-800 rounded mb-4 overflow-hidden">
                  {music.image_path ? (
                    <img
                      src={getFileUrl(music.image_path)}
                      alt={music.song_name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-600">
                      <Play className="w-12 h-12" />
                    </div>
                  )}
                  <button className="absolute bottom-2 right-2 w-12 h-12 bg-green-500 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition shadow-lg hover:scale-105">
                    <Play className="w-6 h-6 text-black ml-1" />
                  </button>
                </div>
                <h3 className="font-semibold truncate">{music.song_name}</h3>
                <p className="text-sm text-gray-400 truncate">
                  {formatDuration(music.duration_seconds)}
                </p>
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
