'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import toast from 'react-hot-toast';
import { useRouter } from 'next/navigation';

interface AlbumFormProps {
  mode: 'create' | 'edit';
  artistUuid: string;
  albumUuid?: string;
  initialName?: string;
  initialDescription?: string;
  onSuccess?: () => void;
}

export default function AlbumForm({
  mode,
  artistUuid,
  albumUuid,
  initialName = '',
  initialDescription = '',
  onSuccess,
}: AlbumFormProps) {
  const router = useRouter();
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);
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
      toast.error('Album name is required');
      return;
    }

    setLoading(true);

    try {
      if (mode === 'create') {
        await api.createAlbum(artistUuid, name, description || undefined, image || undefined);
        toast.success('Album created successfully');
        router.push(`/artists/${artistUuid}`);
      } else if (mode === 'edit' && albumUuid) {
        await api.updateAlbum(albumUuid, name, description || undefined);
        toast.success('Album updated successfully');
        onSuccess?.();
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || `Failed to ${mode} album`);
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
            src={imagePreview || '/default-album.png'}
            alt="Album preview"
            className="w-48 h-48 rounded object-cover mx-auto mb-4"
          />
        </div>
      )}

      {/* Album Name */}
      <div>
        <label className="block text-sm font-semibold mb-2">
          Album Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Enter album name"
          required
        />
      </div>

      {/* Description */}
      <div>
        <label className="block text-sm font-semibold mb-2">Description</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={4}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Describe this album..."
        />
      </div>

      {/* Image Upload (only for create) */}
      {mode === 'create' && (
        <div>
          <label className="block text-sm font-semibold mb-2">Album Cover</label>
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
          ? 'Create Album'
          : 'Update Album'}
      </button>
    </form>
  );
}
