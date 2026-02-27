'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Tag } from '@/lib/types';
import toast from 'react-hot-toast';

interface MusicTagManagerProps {
  musicUuid: string;
}

export default function MusicTagManager({ musicUuid }: MusicTagManagerProps) {
  const [musicTags, setMusicTags] = useState<Tag[]>([]);
  const [allTags, setAllTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddTag, setShowAddTag] = useState(false);
  const [selectedTag, setSelectedTag] = useState('');

  useEffect(() => {
    loadTags();
  }, [musicUuid]);

  const loadTags = async () => {
    try {
      const [musicTagsData, allTagsData] = await Promise.all([
        api.getMusicTags(musicUuid, 100),
        api.getTags(100),
      ]);

      setMusicTags(musicTagsData);
      setAllTags(allTagsData);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load tags');
    } finally {
      setLoading(false);
    }
  };

  const handleAddTag = async () => {
    if (!selectedTag) {
      toast.error('Please select a tag');
      return;
    }

    try {
      await api.assignTagToMusic(musicUuid, selectedTag);
      toast.success('Tag added successfully');
      loadTags();
      setShowAddTag(false);
      setSelectedTag('');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to add tag');
    }
  };

  const handleRemoveTag = async (tagName: string) => {
    try {
      await api.removeTagFromMusic(musicUuid, tagName);
      toast.success('Tag removed successfully');
      loadTags();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to remove tag');
    }
  };

  const availableTags = allTags.filter(
    (tag) => !musicTags.some((mt) => mt.tag_name === tag.tag_name)
  );

  if (loading) {
    return <div className="text-center py-4">Loading tags...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Tags</h3>
        {availableTags.length > 0 && (
          <button
            onClick={() => setShowAddTag(!showAddTag)}
            className="bg-blue-600 hover:bg-blue-700 px-3 py-1 rounded text-sm font-semibold"
          >
            {showAddTag ? 'Cancel' : 'Add Tag'}
          </button>
        )}
      </div>

      {/* Add Tag Form */}
      {showAddTag && availableTags.length > 0 && (
        <div className="bg-gray-700 rounded-lg p-4 space-y-3">
          <select
            value={selectedTag}
            onChange={(e) => setSelectedTag(e.target.value)}
            className="w-full px-4 py-2 bg-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Select a tag</option>
            {availableTags.map((tag) => (
              <option key={tag.tag_name} value={tag.tag_name}>
                {tag.tag_name}
                {tag.tag_description && ` - ${tag.tag_description}`}
              </option>
            ))}
          </select>
          <button
            onClick={handleAddTag}
            className="w-full bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold"
          >
            Add Tag
          </button>
        </div>
      )}

      {/* Current Tags */}
      {musicTags.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {musicTags.map((tag) => (
            <div
              key={tag.tag_name}
              className="flex items-center gap-2 px-3 py-1 bg-blue-600 rounded-full"
            >
              <span className="text-sm">{tag.tag_name}</span>
              <button
                onClick={() => handleRemoveTag(tag.tag_name)}
                className="hover:text-red-400"
                title="Remove tag"
              >
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                    clipRule="evenodd"
                  />
                </svg>
              </button>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-gray-400 text-sm">No tags assigned</p>
      )}
    </div>
  );
}
