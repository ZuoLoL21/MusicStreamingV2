'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import toast from 'react-hot-toast';

interface FollowButtonProps {
  type: 'user' | 'artist';
  uuid: string;
  initialIsFollowing: boolean;
  onFollowChange?: (isFollowing: boolean) => void;
}

export default function FollowButton({
  type,
  uuid,
  initialIsFollowing,
  onFollowChange,
}: FollowButtonProps) {
  const [isFollowing, setIsFollowing] = useState(initialIsFollowing);
  const [loading, setLoading] = useState(false);

  const handleToggle = async () => {
    setLoading(true);
    try {
      if (isFollowing) {
        if (type === 'user') {
          await api.unfollowUser(uuid);
        } else {
          await api.unfollowArtist(uuid);
        }
        toast.success(`Unfollowed ${type}`);
      } else {
        if (type === 'user') {
          await api.followUser(uuid);
        } else {
          await api.followArtist(uuid);
        }
        toast.success(`Following ${type}`);
      }

      const newFollowState = !isFollowing;
      setIsFollowing(newFollowState);
      onFollowChange?.(newFollowState);
    } catch (error: any) {
      toast.error(error.response?.data?.error || `Failed to ${isFollowing ? 'unfollow' : 'follow'} ${type}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleToggle}
      disabled={loading}
      className={`px-4 py-2 rounded-full font-semibold transition disabled:opacity-50 ${
        isFollowing
          ? 'bg-gray-700 hover:bg-gray-600 text-white'
          : 'bg-blue-600 hover:bg-blue-700 text-white'
      }`}
    >
      {loading ? 'Loading...' : isFollowing ? 'Unfollow' : 'Follow'}
    </button>
  );
}
