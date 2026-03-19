'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { SongPopularity } from '@/lib/types';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function PopularPage() {
  const [popularSongs, setPopularSongs] = useState<SongPopularity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPopularSongs();
  }, []);

  const loadPopularSongs = async () => {
    try {
      const songs = await api.getPopularSongsAllTime(50);
      setPopularSongs(songs);
    } catch (error: any) {
      toast.error('Failed to load popular songs');
    } finally {
      setLoading(false);
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
      <h1 className="text-4xl font-bold mb-8">Popular Songs</h1>

      <div className="space-y-2">
        {popularSongs.map((song, index) => (
          <div
            key={song.music_uuid}
            className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition group"
          >
            <span className="text-gray-400 w-8 text-right">{index + 1}</span>
            <div className="flex-1">
              <Link href={`/music/${song.music_uuid}`} className="font-semibold hover:underline">
                {song.song_name}
              </Link>
              <div className="flex items-center gap-2 text-sm text-gray-400">
                <span>{song.artist_name}</span>
              </div>
            </div>
            <div className="text-sm text-gray-400">
              {song.plays?.toLocaleString() || song.decay_plays?.toFixed(0) || '0'} plays
            </div>
            <div className="opacity-0 group-hover:opacity-100 transition">
              <AddToPlaylistButton musicUuid={song.music_uuid} size="sm" />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
