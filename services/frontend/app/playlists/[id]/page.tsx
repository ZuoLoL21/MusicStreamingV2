'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api } from '@/lib/api';
import { Playlist, Music } from '@/lib/types';
import { Play, Music2 } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import toast from 'react-hot-toast';

export default function PlaylistPage() {
  const params = useParams();
  const playlistId = params.id as string;
  const [playlist, setPlaylist] = useState<Playlist | null>(null);
  const [tracks, setTracks] = useState<Music[]>([]);
  const [loading, setLoading] = useState(true);
  const { playQueue } = usePlayerStore();

  useEffect(() => {
    loadPlaylistData();
  }, [playlistId]);

  const loadPlaylistData = async () => {
    try {
      setLoading(true);
      const [playlistData, tracksData] = await Promise.all([
        api.getPlaylist(playlistId),
        api.getPlaylistTracks(playlistId),
      ]);

      setPlaylist(playlistData);
      setTracks(tracksData);
    } catch (error) {
      toast.error('Failed to load playlist');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handlePlayAll = () => {
    if (tracks.length > 0) {
      playQueue(tracks, 0);
    }
  };

  const handlePlayTrack = (track: Music, index: number) => {
    playQueue(tracks, index);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  if (!playlist) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Playlist not found</div>
      </div>
    );
  }

  return (
    <div>
      {/* Playlist Header */}
      <div className="bg-gradient-to-b from-blue-900 to-black p-8">
        <div className="flex items-end space-x-6">
          <div className="w-48 h-48 bg-gradient-to-br from-purple-900 to-blue-900 rounded overflow-hidden flex-shrink-0">
            {playlist.image_path ? (
              <img
                src={playlist.image_path}
                alt={playlist.original_name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <Music2 className="w-20 h-20 text-white opacity-50" />
              </div>
            )}
          </div>
          <div className="flex-1">
            <p className="text-sm font-semibold uppercase">
              {playlist.is_public ? 'Public' : 'Private'} Playlist
            </p>
            <h1 className="text-6xl font-bold my-2">{playlist.original_name}</h1>
            {playlist.description && (
              <p className="text-gray-400 mt-2">{playlist.description}</p>
            )}
            <p className="text-sm text-gray-400 mt-2">
              {tracks.length} {tracks.length === 1 ? 'song' : 'songs'}
            </p>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="bg-gradient-to-b from-black/60 to-black p-8">
        <button
          onClick={handlePlayAll}
          disabled={tracks.length === 0}
          className="w-14 h-14 bg-green-500 rounded-full flex items-center justify-center hover:scale-105 transition disabled:opacity-50"
        >
          <Play className="w-7 h-7 text-black ml-1" />
        </button>
      </div>

      {/* Track List */}
      <div className="p-8">
        {tracks.length === 0 ? (
          <div className="text-center text-gray-400">
            <Music2 className="w-16 h-16 mx-auto mb-4 opacity-50" />
            <p>This playlist is empty</p>
          </div>
        ) : (
          <div className="space-y-2">
            {tracks.map((track, index) => (
              <div
                key={track.uuid}
                onClick={() => handlePlayTrack(track, index)}
                className="flex items-center space-x-4 p-3 hover:bg-gray-900 rounded-lg cursor-pointer group"
              >
                <span className="text-gray-400 w-6 text-center">{index + 1}</span>
                <div className="w-12 h-12 bg-gray-800 rounded overflow-hidden flex-shrink-0">
                  {track.image_path ? (
                    <img
                      src={track.image_path}
                      alt={track.song_name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-600">
                      <Play className="w-6 h-6" />
                    </div>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold truncate">{track.song_name}</h3>
                  <p className="text-sm text-gray-400">Artist Name</p>
                </div>
                <div className="text-sm text-gray-400">
                  {Math.floor(track.duration_seconds / 60)}:{String(track.duration_seconds % 60).padStart(2, '0')}
                </div>
                <button className="w-10 h-10 bg-green-500 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition">
                  <Play className="w-5 h-5 text-black ml-1" />
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
