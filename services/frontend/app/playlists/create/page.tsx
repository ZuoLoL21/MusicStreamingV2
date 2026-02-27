'use client';

import PlaylistForm from '@/components/PlaylistForm';
import Link from 'next/link';

export default function CreatePlaylistPage() {
  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-6">
        <Link href="/profile" className="text-blue-500 hover:underline">
          ← Back to Profile
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg p-6">
        <h1 className="text-3xl font-bold mb-6">Create New Playlist</h1>
        <PlaylistForm mode="create" />
      </div>
    </div>
  );
}
