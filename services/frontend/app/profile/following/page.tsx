'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { User, Artist } from '@/lib/types';
import UserCard from '@/components/UserCard';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function FollowingPage() {
  const [tab, setTab] = useState<'users' | 'artists'>('users');
  const [users, setUsers] = useState<User[]>([]);
  const [artists, setArtists] = useState<Artist[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentUserId, setCurrentUserId] = useState<string>('');

  useEffect(() => {
    loadFollowing();
  }, []);

  const loadFollowing = async () => {
    try {
      const currentUser = await api.getCurrentUser();
      setCurrentUserId(currentUser.uuid);

      const [usersData, artistsData] = await Promise.all([
        api.getFollowingUsers(currentUser.uuid, 50),
        api.getFollowedArtists(currentUser.uuid, 50),
      ]);

      setUsers(usersData);
      setArtists(artistsData);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load following');
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
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Following</h1>

      {/* Tabs */}
      <div className="flex gap-4 mb-6 border-b border-gray-700">
        <button
          onClick={() => setTab('users')}
          className={`px-4 py-2 font-semibold ${
            tab === 'users'
              ? 'text-blue-500 border-b-2 border-blue-500'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Users ({users.length})
        </button>
        <button
          onClick={() => setTab('artists')}
          className={`px-4 py-2 font-semibold ${
            tab === 'artists'
              ? 'text-blue-500 border-b-2 border-blue-500'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Artists ({artists.length})
        </button>
      </div>

      {/* Users Tab */}
      {tab === 'users' && (
        <div>
          {users.length === 0 ? (
            <div className="bg-gray-800 rounded-lg p-8 text-center">
              <p className="text-gray-400">You're not following any users yet</p>
            </div>
          ) : (
            <div className="space-y-4">
              {users.map((user) => (
                <UserCard
                  key={user.uuid}
                  user={user}
                  showBio
                  showFollowButton={user.uuid !== currentUserId}
                />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Artists Tab */}
      {tab === 'artists' && (
        <div>
          {artists.length === 0 ? (
            <div className="bg-gray-800 rounded-lg p-8 text-center">
              <p className="text-gray-400">You're not following any artists yet</p>
              <Link
                href="/artists"
                className="mt-4 inline-block bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold"
              >
                Browse Artists
              </Link>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {artists.map((artist) => (
                <Link
                  key={artist.uuid}
                  href={`/artists/${artist.uuid}`}
                  className="flex items-center gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition"
                >
                  <img
                    src={artist.profile_image_path || '/default-artist.png'}
                    alt={artist.artist_name}
                    className="w-16 h-16 rounded-full object-cover"
                  />
                  <div>
                    <h3 className="font-semibold">{artist.artist_name}</h3>
                    {artist.follower_count !== undefined && (
                      <p className="text-sm text-gray-400">
                        {artist.follower_count} followers
                      </p>
                    )}
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
