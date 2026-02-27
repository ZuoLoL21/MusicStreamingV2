'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Artist } from '@/lib/types';
import { useParams, useRouter } from 'next/navigation';
import AlbumForm from '@/components/AlbumForm';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function CreateAlbumPage() {
  const params = useParams();
  const router = useRouter();
  const artistId = params.id as string;

  const [artist, setArtist] = useState<Artist | null>(null);
  const [loading, setLoading] = useState(true);
  const [canCreate, setCanCreate] = useState(false);

  useEffect(() => {
    loadArtist();
  }, [artistId]);

  const loadArtist = async () => {
    try {
      const artistData = await api.getArtist(artistId);
      setArtist(artistData);

      // Check if user can create albums (Manager+ role)
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(artistId);
        const userMember = members.find((m) => m.user_uuid === currentUser.uuid);

        if (userMember && (userMember.role === 'owner' || userMember.role === 'manager')) {
          setCanCreate(true);
        } else {
          toast.error('You need to be a manager or owner to create albums');
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

  if (!artist || !canCreate) {
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
        <h1 className="text-3xl font-bold mb-2">Create New Album</h1>
        <p className="text-gray-400 mb-6">for {artist.artist_name}</p>
        <AlbumForm mode="create" artistUuid={artistId} />
      </div>
    </div>
  );
}
