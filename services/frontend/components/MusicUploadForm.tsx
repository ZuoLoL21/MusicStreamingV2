'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Album } from '@/lib/types';
import AudioUpload from './AudioUpload';
import toast from 'react-hot-toast';
import { useRouter } from 'next/navigation';

interface MusicUploadFormProps {
  artistUuid: string;
}

export default function MusicUploadForm({ artistUuid }: MusicUploadFormProps) {
  const router = useRouter();
  const [songName, setSongName] = useState('');
  const [audioFile, setAudioFile] = useState<File | null>(null);
  const [duration, setDuration] = useState<number>(0);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [selectedAlbum, setSelectedAlbum] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

  useEffect(() => {
    loadAlbums();
  }, [artistUuid]);

  const loadAlbums = async () => {
    try {
      const albumsData = await api.getArtistAlbums(artistUuid, 100);
      setAlbums(albumsData);
    } catch (error) {
      // Albums are optional
    }
  };

  const handleFileSelect = (file: File, fileDuration: number) => {
    setAudioFile(file);
    setDuration(fileDuration);

    // Auto-fill song name from filename if empty
    if (!songName) {
      const name = file.name.replace(/\.[^/.]+$/, ''); // Remove extension
      setSongName(name);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!songName.trim()) {
      toast.error('Song name is required');
      return;
    }

    if (!audioFile || duration === 0) {
      toast.error('Please select a valid audio file');
      return;
    }

    setLoading(true);
    setUploadProgress(0);

    try {
      // Simulate progress (in production, you'd use axios upload progress)
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => Math.min(prev + 10, 90));
      }, 200);

      await api.uploadMusic(
        artistUuid,
        songName,
        duration,
        audioFile,
        selectedAlbum || undefined
      );

      clearInterval(progressInterval);
      setUploadProgress(100);

      toast.success('Music uploaded successfully');
      router.push(`/artists/${artistUuid}`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to upload music');
    } finally {
      setLoading(false);
      setUploadProgress(0);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Song Name */}
      <div>
        <label className="block text-sm font-semibold mb-2">
          Song Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={songName}
          onChange={(e) => setSongName(e.target.value)}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Enter song name"
          required
        />
      </div>

      {/* Audio File */}
      <AudioUpload onFileSelect={handleFileSelect} maxSizeMB={100} />

      {/* Album Selection */}
      {albums.length > 0 && (
        <div>
          <label className="block text-sm font-semibold mb-2">Album (Optional)</label>
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
      )}

      {/* Upload Progress */}
      {loading && (
        <div>
          <div className="mb-2 flex justify-between text-sm">
            <span>Uploading...</span>
            <span>{uploadProgress}%</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>
        </div>
      )}

      {/* Submit Button */}
      <button
        type="submit"
        disabled={loading || !audioFile}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-3 rounded-lg font-semibold disabled:opacity-50"
      >
        {loading ? 'Uploading...' : 'Upload Music'}
      </button>
    </form>
  );
}
