'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api, getFileUrl } from '@/lib/api';
import { formatDuration } from '@/lib/formatters';
import { Album, Music } from '@/lib/types';
import { Play } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function AlbumPage() {
  const params = useParams();
  const albumId = params.id as string;
  const [album, setAlbum] = useState<Album | null>(null);
  const [tracks, setTracks] = useState<Music[]>([]);
  const [loading, setLoading] = useState(true);
  const { playQueue } = usePlayerStore();

  useEffect(() => {
    loadAlbumData();
  }, [albumId]);

  const loadAlbumData = async () => {
    try {
      setLoading(true);
      const [albumData, tracksData] = await Promise.all([
        api.getAlbum(albumId),
        api.getAlbumMusic(albumId),
      ]);

      setAlbum(albumData);
      setTracks(tracksData);
    } catch (error) {
      toast.error('Failed to load album');
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

  if (!album) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Album not found</div>
      </div>
    );
  }

  return (
    <div>
      {/* Album Header */}
      <div className="bg-gradient-to-b from-purple-900 to-black p-8">
        <div className="flex items-end space-x-6">
          <div className="w-48 h-48 bg-gray-800 rounded overflow-hidden flex-shrink-0">
            {album.image_path ? (
              <img
                src={getFileUrl(album.image_path)}
                alt={album.original_name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-gray-600">
                <Play className="w-20 h-20" />
              </div>
            )}
          </div>
          <div className="flex-1">
            <p className="text-sm font-semibold uppercase">Album</p>
            <h1 className="text-6xl font-bold my-2">{album.original_name}</h1>
            {album.description && (
              <p className="text-gray-400 mt-2">{album.description}</p>
            )}
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
          <div className="text-center text-gray-400">No tracks in this album</div>
        ) : (
          <div className="space-y-2">
            {tracks.map((track, index) => (
              <div
                key={track.uuid}
                className="flex items-center space-x-4 p-3 hover:bg-gray-900 rounded-lg group"
              >
                <span className="text-gray-400 w-6 text-center">{index + 1}</span>
                <div className="w-12 h-12 bg-gray-800 rounded overflow-hidden flex-shrink-0">
                  {track.image_path ? (
                    <img
                      src={getFileUrl(track.image_path)}
                      alt={track.song_name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-gray-600">
                      <Play className="w-6 h-6" />
                    </div>
                  )}
                </div>
                <div className="flex-1 min-w-0 cursor-pointer" onClick={() => handlePlayTrack(track, index)}>
                  <h3 className="font-semibold truncate">{track.song_name}</h3>
                  <p className="text-sm text-gray-400">{track.play_count.toLocaleString()} plays</p>
                </div>
                <div className="text-sm text-gray-400">
                  {formatDuration(track.duration_seconds)}
                </div>
                <div className="opacity-0 group-hover:opacity-100 transition">
                  <AddToPlaylistButton musicUuid={track.uuid} size="sm" />
                </div>
                <button
                  onClick={() => handlePlayTrack(track, index)}
                  className="w-10 h-10 bg-green-500 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition"
                >
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
