'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { ThemeRecommendation, SongPopularity, ArtistPopularity } from '@/lib/types';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function DiscoverPage() {
  const [themeRec, setThemeRec] = useState<ThemeRecommendation | null>(null);
  const [popularSongs, setPopularSongs] = useState<SongPopularity[]>([]);
  const [popularArtists, setPopularArtists] = useState<ArtistPopularity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadRecommendations();
  }, []);

  const loadRecommendations = async () => {
    try {
      // Load each independently to handle partial failures
      const results = await Promise.allSettled([
        api.getThemeRecommendation(),
        api.getPopularSongsAllTime(10),
        api.getPopularArtistsAllTime(10),
      ]);

      if (results[0].status === 'fulfilled') {
        setThemeRec(results[0].value);
      } else {
        console.warn('Failed to load theme recommendation:', results[0].reason);
      }

      if (results[1].status === 'fulfilled') {
        setPopularSongs(results[1].value);
      } else {
        console.warn('Failed to load popular songs:', results[1].reason);
      }

      if (results[2].status === 'fulfilled') {
        setPopularArtists(results[2].value);
      } else {
        console.warn('Failed to load popular artists:', results[2].reason);
      }

      // Only show error if all failed
      const allFailed = results.every(r => r.status === 'rejected');
      if (allFailed) {
        toast.error('Failed to load recommendations. Some services may be unavailable.');
      }
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
      <h1 className="text-4xl font-bold mb-8">Discover</h1>

      {/* Theme Recommendation */}
      {themeRec && (
        <div className="bg-gradient-to-r from-purple-900 to-blue-900 rounded-lg p-6 mb-8">
          <h2 className="text-2xl font-bold mb-2">Recommended Theme for You</h2>
          <p className="text-3xl font-bold text-blue-300 mb-4">{themeRec.recommended_theme}</p>
          <Link
            href={`/discover/themes/${themeRec.recommended_theme}`}
            className="inline-block bg-white text-black px-6 py-2 rounded-full font-semibold hover:bg-gray-200"
          >
            Explore Theme
          </Link>
        </div>
      )}

      {/* Popular Songs */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-bold">Popular Songs</h2>
          <Link href="/discover/popular" className="text-blue-400 hover:underline">
            See All
          </Link>
        </div>
        <div className="space-y-2">
          {popularSongs.map((song, index) => (
            <div
              key={song.music_uuid}
              className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition group"
            >
              <span className="text-gray-400 w-6">{index + 1}</span>
              <div className="flex-1">
                <Link href={`/music/${song.music_uuid}`} className="font-semibold hover:underline">
                  {song.song_name}
                </Link>
                <p className="text-sm text-gray-400">{song.artist_name}</p>
              </div>
              <div className="text-sm text-gray-400">
                {song.plays?.toLocaleString() || song.decay_plays?.toFixed(0)} plays
              </div>
              <div className="opacity-0 group-hover:opacity-100 transition">
                <AddToPlaylistButton musicUuid={song.music_uuid} size="sm" />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Popular Artists */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-2xl font-bold">Popular Artists</h2>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {popularArtists.map((artist) => (
            <Link
              key={artist.artist_uuid}
              href={`/artists/${artist.artist_uuid}`}
              className="bg-gray-800 p-4 rounded-lg hover:bg-gray-700 transition text-center"
            >
              <div className="w-24 h-24 bg-gray-700 rounded-full mx-auto mb-2" />
              <h3 className="font-semibold truncate">{artist.artist_name}</h3>
              <p className="text-xs text-gray-400">
                {artist.plays?.toLocaleString() || artist.decay_plays?.toFixed(0)} plays
              </p>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
}
