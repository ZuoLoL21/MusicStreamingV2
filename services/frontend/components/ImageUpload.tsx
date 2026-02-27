'use client';

import { useState } from 'react';

interface ImageUploadProps {
  currentImage?: string;
  onImageSelect: (file: File) => void;
  onUpload: () => void;
  uploading: boolean;
  maxSizeMB?: number;
  shape?: 'square' | 'circle';
}

export default function ImageUpload({
  currentImage,
  onImageSelect,
  onUpload,
  uploading,
  maxSizeMB = 10,
  shape = 'square',
}: ImageUploadProps) {
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string>('');

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setError('');

    if (!file.type.startsWith('image/')) {
      setError('Please select a valid image file');
      return;
    }

    if (file.size > maxSizeMB * 1024 * 1024) {
      setError(`Image must be less than ${maxSizeMB}MB`);
      return;
    }

    onImageSelect(file);

    const reader = new FileReader();
    reader.onloadend = () => {
      setPreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const imageClasses = shape === 'circle' ? 'rounded-full' : 'rounded';

  return (
    <div className="space-y-4">
      <div className="text-center">
        <img
          src={preview || currentImage || '/default-image.png'}
          alt="Preview"
          className={`w-48 h-48 object-cover mx-auto mb-4 ${imageClasses}`}
        />
      </div>

      <div>
        <label className="block text-sm font-semibold mb-2">Select New Image</label>
        <input
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <p className="text-xs text-gray-400 mt-1">Max {maxSizeMB}MB</p>
      </div>

      {error && <p className="text-red-500 text-sm">{error}</p>}

      <button
        onClick={onUpload}
        disabled={uploading || !preview}
        className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
      >
        {uploading ? 'Uploading...' : 'Upload Image'}
      </button>
    </div>
  );
}
