'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import { Playlist } from '@/lib/types';
import { Music2 } from 'lucide-react';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function LibraryPage() {
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPlaylists();
  }, []);

  const loadPlaylists = async () => {
    try {
      setLoading(true);
      const data = await api.getPlaylists(50);
      setPlaylists(data);
    } catch (error) {
      toast.error('Failed to load playlists');
      console.error(error);
    } finally {
      setLoading(false);
    }
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
      <h1 className="text-4xl font-bold mb-8">Your Library</h1>

      {playlists.length === 0 ? (
        <div className="text-center text-gray-400 mt-12">
          <Music2 className="w-16 h-16 mx-auto mb-4 opacity-50" />
          <p>No playlists yet</p>
          <Link
            href="/playlists/create"
            className="inline-block mt-4 px-6 py-2 bg-green-500 hover:bg-green-600 text-white rounded-full transition"
          >
            Create your first playlist
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
          {playlists.map((playlist) => (
            <Link
              key={playlist.uuid}
              href={`/playlists/${playlist.uuid}`}
              className="bg-gray-900 p-4 rounded-lg hover:bg-gray-800 transition group"
            >
              <div className="aspect-square bg-gradient-to-br from-purple-900 to-blue-900 rounded mb-4 flex items-center justify-center overflow-hidden">
                {playlist.image_path ? (
                  <img
                    src={playlist.image_path}
                    alt={playlist.original_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <Music2 className="w-16 h-16 text-white opacity-50" />
                )}
              </div>
              <h3 className="font-semibold truncate">{playlist.original_name}</h3>
              <p className="text-sm text-gray-400">
                {playlist.is_public ? 'Public' : 'Private'} Playlist
              </p>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
