'use client';

import { User } from '@/lib/types';
import Link from 'next/link';

interface UserCardProps {
  user: User;
  showBio?: boolean;
  showFollowButton?: boolean;
  onFollowToggle?: () => void;
  isFollowing?: boolean;
}

export default function UserCard({
  user,
  showBio = false,
  showFollowButton = false,
  onFollowToggle,
  isFollowing = false,
}: UserCardProps) {
  const profileImage = user.profile_image_path || '/default-avatar.png';

  return (
    <div className="flex items-start gap-4 p-4 bg-gray-800 rounded-lg hover:bg-gray-750 transition">
      <Link href={`/profile/${user.uuid}`}>
        <img
          src={profileImage}
          alt={user.username}
          className="w-16 h-16 rounded-full object-cover cursor-pointer hover:opacity-80"
        />
      </Link>

      <div className="flex-1 min-w-0">
        <Link href={`/profile/${user.uuid}`}>
          <h3 className="text-lg font-semibold hover:underline truncate">
            {user.username}
          </h3>
        </Link>
        {showBio && user.bio && (
          <p className="text-sm text-gray-400 mt-1 line-clamp-2">{user.bio}</p>
        )}
        <p className="text-xs text-gray-500 mt-1">
          Joined {new Date(user.created_at).toLocaleDateString()}
        </p>
      </div>

      {showFollowButton && onFollowToggle && (
        <button
          onClick={onFollowToggle}
          className={`px-4 py-2 rounded-full font-semibold transition ${
            isFollowing
              ? 'bg-gray-700 hover:bg-gray-600 text-white'
              : 'bg-blue-600 hover:bg-blue-700 text-white'
          }`}
        >
          {isFollowing ? 'Unfollow' : 'Follow'}
        </button>
      )}
    </div>
  );
}
