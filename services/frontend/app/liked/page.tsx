'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Music } from '@/lib/types';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { Heart } from 'lucide-react';

export default function LikedSongsPage() {
  const [likedSongs, setLikedSongs] = useState<Music[]>([]);
  const [loading, setLoading] = useState(true);
  const { playQueue } = usePlayerStore();

  useEffect(() => {
    loadLikedSongs();
  }, []);

  const loadLikedSongs = async () => {
    try {
      const currentUser = await api.getCurrentUser();
      const songs = await api.getLikedSongs(currentUser.uuid, 100);
      setLikedSongs(songs);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load liked songs');
    } finally {
      setLoading(false);
    }
  };

  const handlePlay = (song: Music, index: number) => {
    playQueue(likedSongs, index);
  };

  const handlePlayAll = () => {
    if (likedSongs.length > 0) {
      playQueue(likedSongs, 0);
    }
  };

  const handleUnlike = async (songId: string) => {
    try {
      await api.unlikeMusic(songId);
      setLikedSongs(likedSongs.filter((song) => song.uuid !== songId));
      toast.success('Removed from liked songs');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to unlike song');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Loading...</div>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="flex items-center space-x-6 mb-8">
        <div className="w-48 h-48 bg-gradient-to-br from-purple-700 to-blue-900 rounded flex items-center justify-center">
          <Heart className="w-24 h-24 text-white fill-white" />
        </div>
        <div>
          <p className="text-sm font-semibold uppercase">Playlist</p>
          <h1 className="text-6xl font-bold my-2">Liked Songs</h1>
          <p className="text-gray-400">{likedSongs.length} songs</p>
          {likedSongs.length > 0 && (
            <button
              onClick={handlePlayAll}
              className="mt-4 bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded-full font-semibold"
            >
              Play All
            </button>
          )}
        </div>
      </div>

      {likedSongs.length === 0 ? (
        <div className="text-center text-gray-400 mt-12">
          <Heart className="w-16 h-16 mx-auto mb-4 opacity-50" />
          <p>No liked songs yet</p>
          <p className="text-sm mt-2">Songs you like will appear here</p>
          <Link
            href="/search"
            className="mt-4 inline-block bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold text-white"
          >
            Discover Music
          </Link>
        </div>
      ) : (
        <div className="space-y-2">
          {likedSongs.map((song, index) => (
            <div
              key={song.uuid}
              className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition group"
            >
              <button
                onClick={() => handlePlay(song, index)}
                className="text-blue-500 hover:text-blue-400"
              >
                <svg
                  className="w-8 h-8"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="M6.3 2.841A1.5 1.5 0 004 4.11V15.89a1.5 1.5 0 002.3 1.269l9.344-5.89a1.5 1.5 0 000-2.538L6.3 2.84z" />
                </svg>
              </button>

              <img
                src={song.image_path || '/default-music.png'}
                alt={song.song_name}
                className="w-12 h-12 rounded object-cover"
              />

              <div className="flex-1 min-w-0">
                <Link
                  href={`/music/${song.uuid}`}
                  className="font-semibold hover:underline truncate block"
                >
                  {song.song_name}
                </Link>
                <Link
                  href={`/artists/${song.from_artist}`}
                  className="text-sm text-gray-400 hover:underline"
                >
                  Artist
                </Link>
              </div>

              <div className="text-sm text-gray-400">
                {Math.floor(song.duration_seconds / 60)}:
                {(song.duration_seconds % 60).toString().padStart(2, '0')}
              </div>

              <button
                onClick={() => handleUnlike(song.uuid)}
                className="text-red-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition"
                title="Unlike"
              >
                <Heart className="w-6 h-6 fill-current" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
