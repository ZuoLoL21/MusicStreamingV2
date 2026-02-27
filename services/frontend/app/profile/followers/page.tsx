'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { User } from '@/lib/types';
import UserCard from '@/components/UserCard';
import toast from 'react-hot-toast';

export default function FollowersPage() {
  const [followers, setFollowers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentUserId, setCurrentUserId] = useState<string>('');

  useEffect(() => {
    loadFollowers();
  }, []);

  const loadFollowers = async () => {
    try {
      const currentUser = await api.getCurrentUser();
      setCurrentUserId(currentUser.uuid);

      const followersData = await api.getFollowersForUser(currentUser.uuid, 50);
      setFollowers(followersData);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load followers');
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
      <h1 className="text-3xl font-bold mb-6">Your Followers</h1>

      {followers.length === 0 ? (
        <div className="bg-gray-800 rounded-lg p-8 text-center">
          <p className="text-gray-400">You don't have any followers yet</p>
        </div>
      ) : (
        <div className="space-y-4">
          {followers.map((follower) => (
            <UserCard
              key={follower.uuid}
              user={follower}
              showBio
              showFollowButton={follower.uuid !== currentUserId}
            />
          ))}
        </div>
      )}
    </div>
  );
}
