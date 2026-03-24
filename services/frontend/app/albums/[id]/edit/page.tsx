'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Album } from '@/lib/types';
import { useParams, useRouter } from 'next/navigation';
import AlbumForm from '@/components/AlbumForm';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function EditAlbumPage() {
  const params = useParams();
  const router = useRouter();
  const albumId = params.id as string;

  const [album, setAlbum] = useState<Album | null>(null);
  const [loading, setLoading] = useState(true);
  const [canEdit, setCanEdit] = useState(false);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [uploadingImage, setUploadingImage] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    loadAlbum();
  }, [albumId]);

  const loadAlbum = async () => {
    try {
      const albumData = await api.getAlbum(albumId);
      setAlbum(albumData);

      // Check if user can edit (Manager+ for the artist)
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(albumData.from_artist);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);

        if (userMember && (userMember.role === 'owner' || userMember.role === 'manager')) {
          setCanEdit(true);
        } else {
          toast.error('You need to be a manager or owner to edit this album');
          router.push(`/albums/${albumId}`);
        }
      } catch (e) {
        toast.error('Failed to verify permissions');
        router.push(`/albums/${albumId}`);
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load album');
      router.push('/albums');
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
      await api.uploadAlbumImage(albumId, imageFile);
      toast.success('Album cover updated successfully');
      setImageFile(null);
      setImagePreview(null);
      loadAlbum();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to upload image');
    } finally {
      setUploadingImage(false);
    }
  };

  const handleDelete = async () => {
    if (
      !confirm(
        'Are you sure you want to delete this album? This will NOT delete the music in it, but will remove them from the album.'
      )
    ) {
      return;
    }

    setDeleting(true);
    try {
      await api.deleteAlbum(albumId);
      toast.success('Album deleted successfully');
      router.push(`/artists/${album?.from_artist}`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to delete album');
    } finally {
      setDeleting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Loading...</div>
      </div>
    );
  }

  if (!album || !canEdit) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Access denied</div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-6">
        <Link href={`/albums/${albumId}`} className="text-blue-500 hover:underline">
          ← Back to Album
        </Link>
      </div>

      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <h1 className="text-3xl font-bold mb-6">Edit Album</h1>
        <AlbumForm
          mode="edit"
          artistUuid={album.from_artist}
          albumUuid={albumId}
          initialName={album.original_name}
          initialDescription={album.description}
          onSuccess={loadAlbum}
        />
      </div>

      {/* Image Upload Section */}
      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4">Update Album Cover</h2>

        <div className="text-center mb-4">
          <img
            src={imagePreview || album.image_path || '/default-album.png'}
            alt="Album preview"
            className="w-48 h-48 rounded object-cover mx-auto"
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

      {/* Danger Zone */}
      <div className="bg-red-900/20 border border-red-900 rounded-lg p-6">
        <h2 className="text-2xl font-bold mb-2 text-red-400">Danger Zone</h2>
        <p className="text-gray-400 mb-4">
          Deleting this album will not delete the music in it, but will remove them from the
          album.
        </p>
        <button
          onClick={handleDelete}
          disabled={deleting}
          className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
        >
          {deleting ? 'Deleting...' : 'Delete Album'}
        </button>
      </div>
    </div>
  );
}
