'use client';

import { useState, useRef } from 'react';

interface AudioUploadProps {
  onFileSelect: (file: File, duration: number) => void;
  maxSizeMB?: number;
}

export default function AudioUpload({ onFileSelect, maxSizeMB = 100 }: AudioUploadProps) {
  const [file, setFile] = useState<File | null>(null);
  const [duration, setDuration] = useState<number>(0);
  const [error, setError] = useState<string>('');
  const audioRef = useRef<HTMLAudioElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (!selectedFile) return;

    setError('');

    // Check file type
    if (!selectedFile.type.startsWith('audio/')) {
      setError('Please select a valid audio file');
      return;
    }

    // Check file size
    const maxSize = maxSizeMB * 1024 * 1024;
    if (selectedFile.size > maxSize) {
      setError(`File size must be less than ${maxSizeMB}MB`);
      return;
    }

    // Get duration
    const audio = new Audio();
    audio.src = URL.createObjectURL(selectedFile);

    audio.onloadedmetadata = () => {
      const durationSeconds = Math.floor(audio.duration);
      setDuration(durationSeconds);
      setFile(selectedFile);
      onFileSelect(selectedFile, durationSeconds);
      URL.revokeObjectURL(audio.src);
    };

    audio.onerror = () => {
      setError('Failed to load audio file');
      URL.revokeObjectURL(audio.src);
    };
  };

  const formatDuration = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-semibold mb-2">
          Audio File <span className="text-red-500">*</span>
        </label>
        <input
          type="file"
          accept="audio/*"
          onChange={handleFileChange}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-400 mt-1">
          Supported formats: MP3, WAV, FLAC, etc. Max {maxSizeMB}MB
        </p>
      </div>

      {error && <p className="text-red-500 text-sm">{error}</p>}

      {file && duration > 0 && (
        <div className="p-4 bg-gray-700 rounded-lg">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-semibold truncate">{file.name}</p>
              <p className="text-sm text-gray-400">
                {(file.size / (1024 * 1024)).toFixed(2)} MB • {formatDuration(duration)}
              </p>
            </div>
            <svg
              className="w-8 h-8 text-green-500"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                clipRule="evenodd"
              />
            </svg>
          </div>
        </div>
      )}
    </div>
  );
}
