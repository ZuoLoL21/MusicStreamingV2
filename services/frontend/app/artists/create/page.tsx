'use client';

import ArtistForm from '@/components/ArtistForm';
import Link from 'next/link';

export default function CreateArtistPage() {
  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-6">
        <Link href="/profile" className="text-blue-500 hover:underline">
          ← Back to Profile
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg p-6">
        <h1 className="text-3xl font-bold mb-6">Create New Artist</h1>
        <ArtistForm mode="create" />
      </div>

      <div className="mt-6 p-4 bg-gray-800 rounded-lg">
        <h3 className="font-semibold mb-2">What happens next?</h3>
        <ul className="text-sm text-gray-400 space-y-1 list-disc list-inside">
          <li>You'll become the owner of this artist</li>
          <li>You can upload music for this artist</li>
          <li>You can add other members to help manage the artist</li>
          <li>You can create albums and organize your music</li>
        </ul>
      </div>
    </div>
  );
}
