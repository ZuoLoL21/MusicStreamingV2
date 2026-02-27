'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { ListeningHistory, TopMusic } from '@/lib/types';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function HistoryPage() {
  const [tab, setTab] = useState<'recent' | 'top'>('recent');
  const [history, setHistory] = useState<ListeningHistory[]>([]);
  const [topMusic, setTopMusic] = useState<TopMusic[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [tab]);

  const loadData = async () => {
    setLoading(true);
    try {
      if (tab === 'recent') {
        const data = await api.getListeningHistory(50);
        setHistory(data);
      } else {
        const data = await api.getTopMusicForUser(50);
        setTopMusic(data);
      }
    } catch (error: any) {
      toast.error('Failed to load history');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8">
      <h1 className="text-4xl font-bold mb-6">Listening History</h1>

      {/* Tabs */}
      <div className="flex gap-4 mb-6 border-b border-gray-700">
        <button
          onClick={() => setTab('recent')}
          className={`px-4 py-2 font-semibold ${
            tab === 'recent'
              ? 'text-blue-500 border-b-2 border-blue-500'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Recently Played
        </button>
        <button
          onClick={() => setTab('top')}
          className={`px-4 py-2 font-semibold ${
            tab === 'top'
              ? 'text-blue-500 border-b-2 border-blue-500'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Top Tracks
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12">Loading...</div>
      ) : tab === 'recent' ? (
        <div className="space-y-2">
          {history.map((item) => (
            <div
              key={`${item.music_uuid}-${item.listened_at}`}
              className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition"
            >
              <div className="flex-1">
                <Link href={`/music/${item.music_uuid}`} className="font-semibold hover:underline">
                  {item.song_name}
                </Link>
                <div className="flex gap-2 text-sm text-gray-400">
                  <Link href={`/artists/${item.artist_uuid}`} className="hover:underline">
                    {item.artist_name}
                  </Link>
                  <span>•</span>
                  <span>{new Date(item.listened_at).toLocaleString()}</span>
                </div>
              </div>
              {item.completion_percentage !== undefined && (
                <div className="text-sm text-gray-400">{item.completion_percentage}% played</div>
              )}
            </div>
          ))}
          {history.length === 0 && (
            <div className="text-center py-12 text-gray-400">No listening history yet</div>
          )}
        </div>
      ) : (
        <div className="space-y-2">
          {topMusic.map((item, index) => (
            <div
              key={item.music_uuid}
              className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition"
            >
              <span className="text-gray-400 w-6">{index + 1}</span>
              <div className="flex-1">
                <Link href={`/music/${item.music_uuid}`} className="font-semibold hover:underline">
                  {item.song_name}
                </Link>
                <Link
                  href={`/artists/${item.artist_uuid}`}
                  className="block text-sm text-gray-400 hover:underline"
                >
                  {item.artist_name}
                </Link>
              </div>
              <div className="text-sm text-gray-400">{item.play_count} plays</div>
            </div>
          ))}
          {topMusic.length === 0 && (
            <div className="text-center py-12 text-gray-400">No top tracks yet</div>
          )}
        </div>
      )}
    </div>
  );
}
