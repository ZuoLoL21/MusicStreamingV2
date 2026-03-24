'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Artist, ArtistMember } from '@/lib/types';
import { useParams, useRouter } from 'next/navigation';
import MemberManagement from '@/components/MemberManagement';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function ArtistMembersPage() {
  const params = useParams();
  const router = useRouter();
  const artistId = params.id as string;

  const [artist, setArtist] = useState<Artist | null>(null);
  const [currentUserRole, setCurrentUserRole] = useState<'owner' | 'manager' | 'member' | null>(
    null
  );
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadArtist();
  }, [artistId]);

  const loadArtist = async () => {
    try {
      const artistData = await api.getArtist(artistId);
      setArtist(artistData);

      // Get current user's role
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(artistId);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);

        if (userMember) {
          setCurrentUserRole(userMember.role);
        } else {
          // Not a member
          toast.error('You are not a member of this artist');
          router.push(`/artists/${artistId}`);
        }
      } catch (e) {
        toast.error('Failed to verify membership');
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

  if (!artist || !currentUserRole) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Access denied</div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="mb-6">
        <Link href={`/artists/${artistId}`} className="text-blue-500 hover:underline">
          ← Back to Artist
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg p-6">
        <div className="mb-6">
          <h1 className="text-3xl font-bold">{artist.artist_name}</h1>
          <p className="text-gray-400">Manage artist members and roles</p>
          {currentUserRole && (
            <p className="text-sm text-blue-400 mt-2">Your role: {currentUserRole}</p>
          )}
        </div>

        <MemberManagement artistUuid={artistId} currentUserRole={currentUserRole} />

        {currentUserRole === 'owner' && (
          <div className="mt-6 p-4 bg-gray-700 rounded-lg">
            <h3 className="font-semibold mb-2">Role Permissions</h3>
            <ul className="text-sm text-gray-400 space-y-1">
              <li>
                <strong>Owner:</strong> Full control - manage members, edit artist, upload music,
                delete content
              </li>
              <li>
                <strong>Manager:</strong> Edit artist details, upload music, create albums
              </li>
              <li>
                <strong>Member:</strong> Upload music only
              </li>
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}
