'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Artist } from '@/lib/types';
import { useParams, useRouter } from 'next/navigation';
import MusicUploadForm from '@/components/MusicUploadForm';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function UploadMusicPage() {
  const params = useParams();
  const router = useRouter();
  const artistId = params.id as string;

  const [artist, setArtist] = useState<Artist | null>(null);
  const [loading, setLoading] = useState(true);
  const [canUpload, setCanUpload] = useState(false);

  useEffect(() => {
    loadArtist();
  }, [artistId]);

  const loadArtist = async () => {
    try {
      const artistData = await api.getArtist(artistId);
      setArtist(artistData);

      // Check if user can upload (any member role)
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(artistId);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);

        if (userMember) {
          setCanUpload(true);
        } else {
          toast.error('You must be a member of this artist to upload music');
          router.push(`/artists/${artistId}`);
        }
      } catch (e) {
        toast.error('Failed to verify permissions');
        router.push(`/artists/${artistId}`);
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load artist');
      router.push('/artists');
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

  if (!artist || !canUpload) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Access denied</div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-6">
        <Link href={`/artists/${artistId}`} className="text-blue-500 hover:underline">
          ← Back to {artist.artist_name}
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg p-6">
        <h1 className="text-3xl font-bold mb-2">Upload Music</h1>
        <p className="text-gray-400 mb-6">for {artist.artist_name}</p>
        <MusicUploadForm artistUuid={artistId} />
      </div>

      <div className="mt-6 p-4 bg-gray-800 rounded-lg">
        <h3 className="font-semibold mb-2">Upload Guidelines</h3>
        <ul className="text-sm text-gray-400 space-y-1 list-disc list-inside">
          <li>Supported formats: MP3, WAV, FLAC, and other common audio formats</li>
          <li>Maximum file size: 100MB</li>
          <li>Audio duration is automatically detected</li>
          <li>You can optionally assign the music to an album</li>
          <li>Add tags after uploading to help users discover your music</li>
        </ul>
      </div>
    </div>
  );
}
