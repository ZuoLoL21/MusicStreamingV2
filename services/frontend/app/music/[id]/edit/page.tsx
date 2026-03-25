'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Music, Album } from '@/lib/types';
import { useParams, useRouter } from 'next/navigation';
import MusicTagManager from '@/components/MusicTagManager';
import AudioUpload from '@/components/AudioUpload';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function EditMusicPage() {
  const params = useParams();
  const router = useRouter();
  const musicId = params.id as string;

  const [music, setMusic] = useState<Music | null>(null);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [loading, setLoading] = useState(true);
  const [canEdit, setCanEdit] = useState(false);
  const [userRole, setUserRole] = useState<string | null>(null);

  // Edit form state
  const [songName, setSongName] = useState('');
  const [selectedAlbum, setSelectedAlbum] = useState<string>('');
  const [updating, setUpdating] = useState(false);

  // Audio replacement state
  const [newAudioFile, setNewAudioFile] = useState<File | null>(null);
  const [newDuration, setNewDuration] = useState<number>(0);
  const [updatingAudio, setUpdatingAudio] = useState(false);

  // Delete state
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    loadMusic();
  }, [musicId]);

  const loadMusic = async () => {
    try {
      const musicData = await api.getMusic(musicId);
      setMusic(musicData);
      setSongName(musicData.song_name);
      setSelectedAlbum(musicData.in_album || '');

      // Load albums for this artist
      const albumsData = await api.getArtistAlbums(musicData.from_artist, 100);
      setAlbums(albumsData);

      // Check if user can edit (any member of the artist)
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(musicData.from_artist);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);

        if (userMember) {
          setCanEdit(true);
          setUserRole(userMember.role);
        } else {
          toast.error('You must be a member of this artist to edit this music');
          router.push(`/music/${musicId}`);
        }
      } catch (e) {
        toast.error('Failed to verify permissions');
        router.push(`/music/${musicId}`);
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load music');
      router.push('/search');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateDetails = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!songName.trim()) {
      toast.error('Song name is required');
      return;
    }

    setUpdating(true);
    try {
      await api.updateMusicDetails(musicId, songName, selectedAlbum || undefined);
      toast.success('Music details updated successfully');
      loadMusic();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update music');
    } finally {
      setUpdating(false);
    }
  };

  const handleAudioSelect = (file: File, duration: number) => {
    setNewAudioFile(file);
    setNewDuration(duration);
  };

  const handleUpdateAudio = async () => {
    if (!newAudioFile || newDuration === 0) {
      toast.error('Please select a valid audio file');
      return;
    }

    setUpdatingAudio(true);
    try {
      await api.updateMusicStorage(musicId, newAudioFile, newDuration);
      toast.success('Audio file updated successfully');
      setNewAudioFile(null);
      setNewDuration(0);
      loadMusic();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update audio');
    } finally {
      setUpdatingAudio(false);
    }
  };

  const handleDelete = async () => {
    if (
      !confirm(
        'Are you sure you want to delete this music? This action cannot be undone.'
      )
    ) {
      return;
    }

    setDeleting(true);
    try {
      await api.deleteMusic(musicId);
      toast.success('Music deleted successfully');
      router.push(`/artists/${music?.from_artist}`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to delete music');
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

  if (!music || !canEdit) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Access denied</div>
      </div>
    );
  }

  const canDelete = userRole === 'owner';

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="mb-6">
        <Link href={`/music/${musicId}`} className="text-blue-500 hover:underline">
          ← Back to Music
        </Link>
      </div>

      {/* Edit Details */}
      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <h1 className="text-3xl font-bold mb-6">Edit Music Details</h1>
        <form onSubmit={handleUpdateDetails} className="space-y-4">
          <div>
            <label className="block text-sm font-semibold mb-2">
              Song Name <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={songName}
              onChange={(e) => setSongName(e.target.value)}
              className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-semibold mb-2">Album</label>
            <select
              value={selectedAlbum}
              onChange={(e) => setSelectedAlbum(e.target.value)}
              className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">None (Single)</option>
              {albums.map((album) => (
                <option key={album.uuid} value={album.uuid}>
                  {album.original_name}
                </option>
              ))}
            </select>
          </div>

          <button
            type="submit"
            disabled={updating}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
          >
            {updating ? 'Updating...' : 'Update Details'}
          </button>
        </form>
      </div>

      {/* Tag Management */}
      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <MusicTagManager musicUuid={musicId} />
      </div>

      {/* Replace Audio File */}
      <div className="bg-gray-800 rounded-lg p-6 mb-6">
        <h2 className="text-2xl font-bold mb-4">Replace Audio File</h2>
        <p className="text-gray-400 text-sm mb-4">
          Upload a new audio file to replace the current one. Duration will be updated
          automatically.
        </p>
        <div className="space-y-4">
          <AudioUpload onFileSelect={handleAudioSelect} maxSizeMB={100} />
          <button
            onClick={handleUpdateAudio}
            disabled={updatingAudio || !newAudioFile}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
          >
            {updatingAudio ? 'Uploading...' : 'Replace Audio File'}
          </button>
        </div>
      </div>

      {/* Danger Zone (Owner only) */}
      {canDelete && (
        <div className="bg-red-900/20 border border-red-900 rounded-lg p-6">
          <h2 className="text-2xl font-bold mb-2 text-red-400">Danger Zone</h2>
          <p className="text-gray-400 mb-4">
            Deleting this music will permanently remove it from the platform. This action
            cannot be undone.
          </p>
          <button
            onClick={handleDelete}
            disabled={deleting}
            className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
          >
            {deleting ? 'Deleting...' : 'Delete Music'}
          </button>
        </div>
      )}
    </div>
  );
}
