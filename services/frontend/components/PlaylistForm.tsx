'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import toast from 'react-hot-toast';
import { useRouter } from 'next/navigation';

interface PlaylistFormProps {
  mode: 'create' | 'edit';
  playlistUuid?: string;
  initialName?: string;
  initialDescription?: string;
  initialIsPublic?: boolean;
  onSuccess?: () => void;
}

export default function PlaylistForm({
  mode,
  playlistUuid,
  initialName = '',
  initialDescription = '',
  initialIsPublic = true,
  onSuccess,
}: PlaylistFormProps) {
  const router = useRouter();
  const [name, setName] = useState(initialName);
  const [description, setDescription] = useState(initialDescription);
  const [isPublic, setIsPublic] = useState(initialIsPublic);
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
      reader.onloadend = () => setImagePreview(reader.result as string);
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Playlist name is required');
      return;
    }

    setLoading(true);
    try {
      if (mode === 'create') {
        await api.createPlaylist(name, description || undefined, isPublic, image || undefined);
        toast.success('Playlist created successfully');
        router.push('/profile');
      } else if (mode === 'edit' && playlistUuid) {
        await api.updatePlaylist(playlistUuid, name, description || undefined, isPublic);
        toast.success('Playlist updated successfully');
        onSuccess?.();
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || `Failed to ${mode} playlist`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {(imagePreview || mode === 'edit') && (
        <div className="text-center">
          <img
            src={imagePreview || '/default-playlist.png'}
            alt="Playlist preview"
            className="w-48 h-48 rounded object-cover mx-auto mb-4"
          />
        </div>
      )}

      <div>
        <label className="block text-sm font-semibold mb-2">
          Playlist Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="My Awesome Playlist"
          required
        />
      </div>

      <div>
        <label className="block text-sm font-semibold mb-2">Description</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Describe your playlist..."
        />
      </div>

      <div className="flex items-center gap-3">
        <input
          type="checkbox"
          id="isPublic"
          checked={isPublic}
          onChange={(e) => setIsPublic(e.target.checked)}
          className="w-4 h-4"
        />
        <label htmlFor="isPublic" className="text-sm">
          Public (visible to everyone)
        </label>
      </div>

      {mode === 'create' && (
        <div>
          <label className="block text-sm font-semibold mb-2">Cover Image</label>
          <input
            type="file"
            accept="image/*"
            onChange={handleImageSelect}
            className="w-full px-4 py-2 bg-gray-700 rounded-lg"
          />
          <p className="text-xs text-gray-400 mt-1">Max 10MB</p>
        </div>
      )}

      <button
        type="submit"
        disabled={loading}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-3 rounded-lg font-semibold disabled:opacity-50"
      >
        {loading ? (mode === 'create' ? 'Creating...' : 'Updating...') : mode === 'create' ? 'Create Playlist' : 'Update Playlist'}
      </button>
    </form>
  );
}
