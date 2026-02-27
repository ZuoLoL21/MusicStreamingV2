'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import toast from 'react-hot-toast';
import { useRouter } from 'next/navigation';

interface ArtistFormProps {
  mode: 'create' | 'edit';
  artistUuid?: string;
  initialName?: string;
  initialBio?: string;
  onSuccess?: () => void;
}

export default function ArtistForm({
  mode,
  artistUuid,
  initialName = '',
  initialBio = '',
  onSuccess,
}: ArtistFormProps) {
  const router = useRouter();
  const [name, setName] = useState(initialName);
  const [bio, setBio] = useState(initialBio);
  const [image, setImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (file.size > 10 * 1024 * 1024) {
        toast.error('Image must be less than 10MB');
        return;
      }
      setImage(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error('Artist name is required');
      return;
    }

    setLoading(true);

    try {
      if (mode === 'create') {
        await api.createArtist(name, bio || undefined, image || undefined);
        toast.success('Artist created successfully');
        router.push('/profile');
      } else if (mode === 'edit' && artistUuid) {
        await api.updateArtistProfile(artistUuid, name, bio || undefined);
        toast.success('Artist updated successfully');
        onSuccess?.();
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || `Failed to ${mode} artist`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Image Preview */}
      {(imagePreview || mode === 'edit') && (
        <div className="text-center">
          <img
            src={imagePreview || '/default-artist.png'}
            alt="Artist preview"
            className="w-32 h-32 rounded-full object-cover mx-auto mb-4"
          />
        </div>
      )}

      {/* Artist Name */}
      <div>
        <label className="block text-sm font-semibold mb-2">
          Artist Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Enter artist name"
          required
        />
      </div>

      {/* Bio */}
      <div>
        <label className="block text-sm font-semibold mb-2">Bio</label>
        <textarea
          value={bio}
          onChange={(e) => setBio(e.target.value)}
          rows={4}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Tell us about this artist..."
        />
      </div>

      {/* Image Upload (only for create) */}
      {mode === 'create' && (
        <div>
          <label className="block text-sm font-semibold mb-2">Profile Image</label>
          <input
            type="file"
            accept="image/*"
            onChange={handleImageSelect}
            className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <p className="text-xs text-gray-400 mt-1">Max 10MB</p>
        </div>
      )}

      {/* Submit Button */}
      <button
        type="submit"
        disabled={loading}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-3 rounded-lg font-semibold disabled:opacity-50"
      >
        {loading
          ? mode === 'create'
            ? 'Creating...'
            : 'Updating...'
          : mode === 'create'
          ? 'Create Artist'
          : 'Update Artist'}
      </button>
    </form>
  );
}
