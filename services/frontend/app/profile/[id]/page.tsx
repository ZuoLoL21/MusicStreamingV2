'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { User, Artist } from '@/lib/types';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import FollowButton from '@/components/FollowButton';
import toast from 'react-hot-toast';

export default function PublicProfilePage() {
  const params = useParams();
  const userId = params.id as string;

  const [user, setUser] = useState<User | null>(null);
  const [artists, setArtists] = useState<Artist[]>([]);
  const [loading, setLoading] = useState(true);
  const [isOwnProfile, setIsOwnProfile] = useState(false);
  const [isFollowing, setIsFollowing] = useState(false);

  useEffect(() => {
    loadProfile();
  }, [userId]);

  const loadProfile = async () => {
    try {
      const userData = await api.getUser(userId);
      setUser(userData);

      // Check if this is the current user's profile
      try {
        const currentUser = await api.getCurrentUser();
        setIsOwnProfile(currentUser.uuid === userId);
      } catch (e) {
        setIsOwnProfile(false);
      }

      // Load user's artists
      try {
        const artistsData = await api.getUserArtists(userId);
        setArtists(artistsData);
      } catch (e) {
        // Artists might not be public
      }

      // TODO: Check if following (would need an endpoint)
      setIsFollowing(false);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load profile');
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

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">User not found</div>
      </div>
    );
  }

  const profileImage = user.profile_image_path || '/default-avatar.png';

  return (
    <div className="max-w-4xl mx-auto p-6">
      {/* Profile Header */}
      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <div className="flex items-start gap-6">
          <img
            src={profileImage}
            alt={user.username}
            className="w-32 h-32 rounded-full object-cover"
          />
          <div className="flex-1">
            <div className="flex items-center justify-between mb-2">
              <h1 className="text-3xl font-bold">{user.username}</h1>
              {!isOwnProfile && (
                <FollowButton
                  type="user"
                  uuid={userId}
                  initialIsFollowing={isFollowing}
                  onFollowChange={setIsFollowing}
                />
              )}
              {isOwnProfile && (
                <Link
                  href="/profile"
                  className="bg-gray-700 hover:bg-gray-600 px-4 py-2 rounded-lg font-semibold"
                >
                  View My Profile
                </Link>
              )}
            </div>
            {user.bio && <p className="text-gray-300 mb-4">{user.bio}</p>}
            <p className="text-xs text-gray-500">
              Joined {new Date(user.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>
      </div>

      {/* User's Artists */}
      {artists.length > 0 && (
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-2xl font-bold mb-4">Artists</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {artists.map((artist) => (
              <Link
                key={artist.uuid}
                href={`/artists/${artist.uuid}`}
                className="flex items-center gap-4 p-4 bg-gray-700 rounded-lg hover:bg-gray-600 transition"
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
        </div>
      )}
    </div>
  );
}
