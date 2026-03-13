'use client';

import { useState, useEffect } from 'react';
import { api, getFileUrl } from '@/lib/api';
import { User, Artist } from '@/lib/types';
import ProfileEdit from '@/components/ProfileEdit';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function ProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [artists, setArtists] = useState<Artist[]>([]);
  const [loading, setLoading] = useState(true);
  const [showEdit, setShowEdit] = useState(false);
  const [stats, setStats] = useState({
    followers: 0,
    following: 0,
    likedSongs: 0,
  });

  const loadProfile = async () => {
    try {
      const userData = await api.getCurrentUser();
      setUser(userData);

      // Load user's artists
      const artistsData = await api.getUserArtists(userData.uuid);
      setArtists(artistsData);

      // Load stats (count items by fetching with limit 1 and checking if there are more)
      // For now, just load initial counts
      try {
        const followers = await api.getFollowersForUser(userData.uuid, 1);
        const following = await api.getFollowingUsers(userData.uuid, 1);
        const liked = await api.getLikedSongs(userData.uuid, 1);
        // Note: This doesn't give exact counts, but gives an indication
        // A better approach would be to have count endpoints
      } catch (e) {
        // Stats loading is optional
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load profile');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProfile();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">User not found</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      {/* Profile Header */}
      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <div className="flex items-start gap-6">
          <img
            src={getFileUrl(user.profile_image_path || '')}
            alt={user.username}
            className="w-32 h-32 rounded-full object-cover"
          />
          <div className="flex-1">
            <div className="flex items-center justify-between mb-2">
              <h1 className="text-3xl font-bold">{user.username}</h1>
              <button
                onClick={() => setShowEdit(true)}
                className="bg-gray-700 hover:bg-gray-600 px-4 py-2 rounded-lg font-semibold"
              >
                Edit Profile
              </button>
            </div>
            <p className="text-gray-400 mb-4">{user.email}</p>
            {user.bio && <p className="text-gray-300 mb-4">{user.bio}</p>}
            <div className="flex gap-6 text-sm">
              <Link href="/profile/followers" className="hover:text-blue-400">
                <span className="font-semibold">{stats.followers}</span> Followers
              </Link>
              <Link href="/profile/following" className="hover:text-blue-400">
                <span className="font-semibold">{stats.following}</span> Following
              </Link>
              <Link href="/liked" className="hover:text-blue-400">
                <span className="font-semibold">{stats.likedSongs}</span> Liked Songs
              </Link>
            </div>
            <p className="text-xs text-gray-500 mt-2">
              Joined {new Date(user.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>
      </div>

      {/* User's Artists */}
      {artists.length > 0 && (
        <div className="bg-gray-800 rounded-lg p-6 mb-6">
          <h2 className="text-2xl font-bold mb-4">Your Artists</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {artists.map((artist) => (
              <Link
                key={artist.uuid}
                href={`/artists/${artist.uuid}`}
                className="flex items-center gap-4 p-4 bg-gray-700 rounded-lg hover:bg-gray-600 transition"
              >
                <img
                  src={getFileUrl(artist.profile_image_path || '')}
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
          <Link
            href="/artists/create"
            className="mt-4 inline-block bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold"
          >
            Create New Artist
          </Link>
        </div>
      )}

      {artists.length === 0 && (
        <div className="bg-gray-800 rounded-lg p-6 text-center">
          <p className="text-gray-400 mb-4">You haven't created any artists yet</p>
          <Link
            href="/artists/create"
            className="inline-block bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold"
          >
            Create Your First Artist
          </Link>
        </div>
      )}

      {/* Edit Modal */}
      {showEdit && (
        <ProfileEdit
          user={user}
          onClose={() => setShowEdit(false)}
          onUpdate={() => {
            loadProfile();
            setShowEdit(false);
          }}
        />
      )}
    </div>
  );
}
