'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Artist } from '@/lib/types';
import { useParams, useRouter } from 'next/navigation';
import ArtistForm from '@/components/ArtistForm';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function EditArtistPage() {
  const params = useParams();
  const router = useRouter();
  const artistId = params.id as string;

  const [artist, setArtist] = useState<Artist | null>(null);
  const [loading, setLoading] = useState(true);
  const [canEdit, setCanEdit] = useState(false);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [uploadingImage, setUploadingImage] = useState(false);

  useEffect(() => {
    loadArtist();
  }, [artistId]);

  const loadArtist = async () => {
    try {
      const data = await api.getArtist(artistId);
      setArtist(data);

      // Check if user can edit (owner or manager only)
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(artistId);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);

        if (userMember && (userMember.role === 'owner' || userMember.role === 'manager')) {
          setCanEdit(true);
        } else {
          toast.error('You must be an owner or manager to edit this artist');
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

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        toast.error('Image must be less than 10MB');
        return;
      }
      setImageFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleImageUpload = async () => {
    if (!imageFile) {
      toast.error('Please select an image');
      return;
    }

    setUploadingImage(true);
    try {
      await api.uploadArtistImage(artistId, imageFile);
      toast.success('Artist image updated successfully');
      setImageFile(null);
      setImagePreview(null);
      loadArtist();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to upload image');
    } finally {
      setUploadingImage(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Loading...</div>
      </div>
    );
  }

  if (!artist || !canEdit) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">You don't have permission to edit this artist</div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-6">
        <Link href={`/artists/${artistId}`} className="text-blue-500 hover:underline">
          ← Back to Artist
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <h1 className="text-3xl font-bold mb-6">Edit Artist Profile</h1>
        <ArtistForm
          mode="edit"
          artistUuid={artistId}
          initialName={artist.artist_name}
          initialBio={artist.bio}
          onSuccess={loadArtist}
        />
      </div>

      {/* Image Upload Section */}
      <div className="bg-gray-800 rounded-lg p-6">
        <h2 className="text-2xl font-bold mb-4">Update Profile Image</h2>

        <div className="text-center mb-4">
          <img
            src={imagePreview || artist.profile_image_path || '/default-artist.png'}
            alt="Artist preview"
            className="w-32 h-32 rounded-full object-cover mx-auto"
          />
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-semibold mb-2">Select New Image</label>
            <input
              type="file"
              accept="image/*"
              onChange={handleImageSelect}
              className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <p className="text-xs text-gray-400 mt-1">Max 10MB</p>
          </div>

          <button
            onClick={handleImageUpload}
            disabled={uploadingImage || !imageFile}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
          >
            {uploadingImage ? 'Uploading...' : 'Upload Image'}
          </button>
        </div>
      </div>
    </div>
  );
}
