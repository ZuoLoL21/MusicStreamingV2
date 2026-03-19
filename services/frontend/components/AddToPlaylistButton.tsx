'use client';

import { useState, useEffect, useRef } from 'react';
import { ListPlus, Loader2 } from 'lucide-react';
import { api, getFileUrl } from '@/lib/api';
import { Playlist } from '@/lib/types';
import toast from 'react-hot-toast';
import Link from 'next/link';

interface AddToPlaylistButtonProps {
  musicUuid: string;
  className?: string;
  size?: 'sm' | 'md';
}

export function AddToPlaylistButton({ musicUuid, className = '', size = 'md' }: AddToPlaylistButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [loading, setLoading] = useState(false);
  const [addingTo, setAddingTo] = useState<string | null>(null);
  const [userUuid, setUserUuid] = useState<string | null>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const sizeClasses = size === 'sm' ? 'w-8 h-8' : 'w-10 h-10';
  const iconSize = size === 'sm' ? 'w-4 h-4' : 'w-5 h-5';

  // Close dropdown on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  // Fetch playlists when dropdown opens
  useEffect(() => {
    const fetchPlaylists = async () => {
      if (isOpen && playlists.length === 0) {
        setLoading(true);
        try {
          const user = await api.getCurrentUser();
          setUserUuid(user.uuid);
          const userPlaylists = await api.getUserPlaylists(user.uuid, 50);
          setPlaylists(userPlaylists);
        } catch (error) {
          toast.error('Failed to load playlists');
          console.error(error);
        } finally {
          setLoading(false);
        }
      }
    };

    fetchPlaylists();
  }, [isOpen, playlists.length]);

  const handleAddToPlaylist = async (playlist: Playlist) => {
    setAddingTo(playlist.uuid);
    try {
      // Get track count to determine position
      const tracks = await api.getPlaylistTracks(playlist.uuid);
      const position = tracks.length + 1;

      await api.addTrackToPlaylist(playlist.uuid, musicUuid, position);
      toast.success(`Added to ${playlist.original_name}`);
      setIsOpen(false);
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'Failed to add to playlist';
      toast.error(errorMessage);
      console.error(error);
    } finally {
      setAddingTo(null);
    }
  };

  const toggleDropdown = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsOpen(!isOpen);
  };

  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      <button
        onClick={toggleDropdown}
        className={`${sizeClasses} bg-gray-800 hover:bg-gray-700 rounded-full flex items-center justify-center transition`}
        title="Add to playlist"
      >
        <ListPlus className={iconSize} />
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-64 bg-gray-900 border border-gray-700 rounded-lg shadow-lg z-50 max-h-80 overflow-y-auto">
          {loading ? (
            <div className="p-4 flex items-center justify-center">
              <Loader2 className="w-5 h-5 animate-spin" />
            </div>
          ) : playlists.length === 0 ? (
            <div className="p-4">
              <p className="text-gray-400 text-sm mb-3 text-center">No playlists yet</p>
              <Link
                href="/playlists/create"
                className="block w-full py-2 px-3 bg-green-500 hover:bg-green-600 rounded text-center font-semibold transition"
                onClick={() => setIsOpen(false)}
              >
                Create Playlist
              </Link>
            </div>
          ) : (
            <div className="py-2">
              {playlists.map((playlist) => (
                <button
                  key={playlist.uuid}
                  onClick={() => handleAddToPlaylist(playlist)}
                  disabled={addingTo === playlist.uuid}
                  className="w-full px-3 py-2 hover:bg-gray-800 flex items-center gap-3 transition disabled:opacity-50"
                >
                  <div className="w-10 h-10 bg-gray-800 rounded overflow-hidden flex-shrink-0">
                    {playlist.image_path ? (
                      <img
                        src={getFileUrl(playlist.image_path)}
                        alt={playlist.original_name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-gray-600">
                        <ListPlus className="w-5 h-5" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1 text-left min-w-0">
                    <p className="font-semibold truncate">{playlist.original_name}</p>
                    <p className="text-xs text-gray-400">
                      {playlist.is_public ? 'Public' : 'Private'}
                    </p>
                  </div>
                  {addingTo === playlist.uuid && (
                    <Loader2 className="w-4 h-4 animate-spin flex-shrink-0" />
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
