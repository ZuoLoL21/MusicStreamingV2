'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { api } from '@/lib/api';
import { SongPopularity } from '@/lib/types';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function ThemePage() {
  const params = useParams();
  const theme = decodeURIComponent(params.theme as string);
  const [songs, setSongs] = useState<SongPopularity[]>([]);
  const [loading, setLoading] = useState(true);
  const { playTrack } = usePlayerStore();

  useEffect(() => {
    loadThemeSongs();
  }, [theme]);

  const loadThemeSongs = async () => {
    try {
      setLoading(true);
      const data = await api.getPopularSongsByTheme(theme, 50);
      setSongs(data);
    } catch (error: any) {
      console.error('Failed to load theme songs:', error);
      if (error.response?.status === 404) {
        toast.error(`Theme "${theme}" not found`);
      } else {
        toast.error('Failed to load songs for this theme');
      }
    } finally {
      setLoading(false);
    }
  };

  const handlePlaySong = async (song: SongPopularity) => {
    try {
      // Fetch full song details to get the audio path
      const songDetails = await api.getMusic(song.music_uuid);
      playTrack(songDetails);
    } catch (error) {
      console.error('Failed to load song details:', error);
      toast.error('Failed to play song');
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
      <div className="mb-8">
        <Link
          href="/discover"
          className="text-blue-400 hover:underline mb-4 inline-block"
        >
          &larr; Back to Discover
        </Link>
        <h1 className="text-4xl font-bold mb-2">{theme}</h1>
        <p className="text-gray-400">
          {songs.length > 0
            ? `${songs.length} popular songs in this theme`
            : 'No songs available yet'}
        </p>
      </div>

      {songs.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-xl text-gray-400 mb-4">
            No songs found for this theme yet.
          </p>
          <p className="text-gray-500">
            Start listening to music to build up recommendations!
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {songs.map((song, index) => (
            <div
              key={song.music_uuid}
              className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition group"
            >
              <span className="text-gray-400 w-8 text-right">{index + 1}</span>
              <button
                onClick={() => handlePlaySong(song)}
                className="w-10 h-10 flex items-center justify-center bg-blue-500 hover:bg-blue-600 rounded-full transition"
                aria-label="Play"
              >
                <svg
                  className="w-5 h-5 ml-0.5"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="M6.3 2.841A1.5 1.5 0 004 4.11V15.89a1.5 1.5 0 002.3 1.269l9.344-5.89a1.5 1.5 0 000-2.538L6.3 2.84z" />
                </svg>
              </button>
              <div className="flex-1 min-w-0">
                <Link
                  href={`/music/${song.music_uuid}`}
                  className="font-semibold hover:underline block truncate"
                >
                  {song.song_name || 'Unknown Song'}
                </Link>
                <Link
                  href={`/artists/${song.artist_uuid}`}
                  className="text-sm text-gray-400 hover:underline block truncate"
                >
                  {song.artist_name || 'Unknown Artist'}
                </Link>
              </div>
              <div className="text-sm text-gray-400">
                {song.plays?.toLocaleString() || song.decay_plays?.toFixed(0)}{' '}
                plays
              </div>
              <div className="opacity-0 group-hover:opacity-100 transition">
                <AddToPlaylistButton musicUuid={song.music_uuid} size="sm" />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
