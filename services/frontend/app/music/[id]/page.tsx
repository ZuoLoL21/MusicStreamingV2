'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api, getFileUrl } from '@/lib/api';
import { formatDuration } from '@/lib/formatters';
import { Music, Tag } from '@/lib/types';
import { Play, Heart, Edit } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function MusicPage() {
  const params = useParams();
  const musicId = params.id as string;
  const [music, setMusic] = useState<Music | null>(null);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [liked, setLiked] = useState(false);
  const [canEdit, setCanEdit] = useState(false);
  const { playQueue } = usePlayerStore();

  useEffect(() => {
    loadMusicData();
  }, [musicId]);

  const loadMusicData = async () => {
    try {
      setLoading(true);
      const [musicData, tagsData] = await Promise.all([
        api.getMusic(musicId),
        api.getMusicTags(musicId, 50),
      ]);

      setMusic(musicData);
      setTags(tagsData);

      // Check if liked
      try {
        const likedResponse = await api.checkIfMusicLiked(musicId);
        setLiked(likedResponse.liked);
      } catch (e) {
        // User not logged in or error checking like status
      }

      // Check if user can edit (any member of the artist)
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(musicData.from_artist);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);
        if (userMember) {
          setCanEdit(true);
        }
      } catch (e) {
        // User not logged in or not a member
      }
    } catch (error) {
      toast.error('Failed to load music');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handlePlay = () => {
    if (music) {
      playQueue([music], 0);
    }
  };

  const handleLike = async () => {
    if (!music) return;

    try {
      if (liked) {
        await api.unlikeMusic(musicId);
        setLiked(false);
        toast.success('Removed from liked songs');
      } else {
        await api.likeMusic(musicId);
        setLiked(true);
        toast.success('Added to liked songs');
        // Send like event for recommendation system
        await api.sendLikeEvent(music.uuid, music.from_artist);
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update like status');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  if (!music) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Music not found</div>
      </div>
    );
  }

  return (
    <div>
      {/* Music Header */}
      <div className="bg-gradient-to-b from-blue-900 to-black p-8">
        <div className="flex items-end space-x-6">
          <div className="w-48 h-48 bg-gray-800 rounded overflow-hidden flex-shrink-0">
            {music.image_path ? (
              <img
                src={getFileUrl(music.image_path)}
                alt={music.song_name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-gray-600">
                <Play className="w-20 h-20" />
              </div>
            )}
          </div>
          <div className="flex-1">
            <p className="text-sm font-semibold uppercase">Song</p>
            <h1 className="text-6xl font-bold my-2">{music.song_name}</h1>
            <div className="flex items-center gap-2 text-gray-400">
              <Link href={`/artists/${music.from_artist}`} className="hover:underline">
                Artist
              </Link>
              {music.in_album && (
                <>
                  <span>•</span>
                  <Link href={`/albums/${music.in_album}`} className="hover:underline">
                    Album
                  </Link>
                </>
              )}
              <span>•</span>
              <span>{formatDuration(music.duration_seconds)}</span>
              <span>•</span>
              <span>{music.play_count.toLocaleString()} plays</span>
            </div>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="bg-gradient-to-b from-black/60 to-black p-8">
        <div className="flex items-center gap-4">
          <button
            onClick={handlePlay}
            className="w-14 h-14 bg-green-500 rounded-full flex items-center justify-center hover:scale-105 transition"
          >
            <Play className="w-7 h-7 text-black ml-1" />
          </button>

          <button
            onClick={handleLike}
            className={`w-12 h-12 rounded-full flex items-center justify-center transition ${
              liked
                ? 'text-green-500 hover:text-green-400'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            <Heart className={`w-7 h-7 ${liked ? 'fill-current' : ''}`} />
          </button>

          <AddToPlaylistButton musicUuid={musicId} />

          {canEdit && (
            <Link
              href={`/music/${musicId}/edit`}
              className="w-12 h-12 rounded-full flex items-center justify-center text-gray-400 hover:text-white transition"
            >
              <Edit className="w-6 h-6" />
            </Link>
          )}
        </div>
      </div>

      {/* Tags */}
      {tags.length > 0 && (
        <div className="p-8">
          <h2 className="text-xl font-bold mb-4">Tags</h2>
          <div className="flex flex-wrap gap-2">
            {tags.map((tag) => (
              <Link
                key={tag.tag_name}
                href={`/tags/${tag.tag_name}`}
                className="px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded-full text-sm transition"
              >
                {tag.tag_name}
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
