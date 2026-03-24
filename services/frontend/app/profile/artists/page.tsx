'use client';

import { useState, useEffect } from 'react';
import { api, getFileUrl } from '@/lib/api';
import { Artist, Music, Album } from '@/lib/types';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { Upload, Edit, Users, Eye, Music as MusicIcon, Disc, UserPlus } from 'lucide-react';

interface ArtistStats {
  trackCount: number;
  albumCount: number;
  followerCount: number;
}

export default function MyArtistsPage() {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [artistsStats, setArtistsStats] = useState<Map<string, ArtistStats>>(new Map());
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadArtists();
  }, []);

  const loadArtists = async () => {
    try {
      const currentUser = await api.getCurrentUser();
      const artistsData = await api.getUserArtists(currentUser.uuid);
      setArtists(artistsData);

      // Load stats for each artist
      const statsMap = new Map<string, ArtistStats>();
      await Promise.all(
        artistsData.map(async (artist) => {
          try {
            const [music, albums] = await Promise.all([
              api.getArtistMusic(artist.uuid, 1000), // Get all tracks
              api.getArtistAlbums(artist.uuid, 1000), // Get all albums
            ]);

            statsMap.set(artist.uuid, {
              trackCount: music.length,
              albumCount: albums.length,
              followerCount: artist.follower_count || 0,
            });
          } catch (error) {
            console.error(`Failed to load stats for artist ${artist.uuid}:`, error);
            statsMap.set(artist.uuid, {
              trackCount: 0,
              albumCount: 0,
              followerCount: artist.follower_count || 0,
            });
          }
        })
      );
      setArtistsStats(statsMap);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load artists');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Loading your artists...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-6">
      {/* Header */}
      <div className="mb-8">
        <Link href="/profile" className="text-blue-500 hover:underline mb-4 inline-block">
          ← Back to Profile
        </Link>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-4xl font-bold mb-2">My Artists</h1>
            <p className="text-gray-400">Manage your artist profiles and content</p>
          </div>
          <Link
            href="/artists/create"
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 px-6 py-3 rounded-lg font-semibold"
          >
            <UserPlus className="w-5 h-5" />
            Create New Artist
          </Link>
        </div>
      </div>

      {/* Artists Grid */}
      {artists.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {artists.map((artist) => {
            const stats = artistsStats.get(artist.uuid);
            return (
              <div
                key={artist.uuid}
                className="bg-gray-800 rounded-lg overflow-hidden hover:bg-gray-750 transition"
              >
                {/* Artist Header */}
                <div className="relative h-32 bg-gradient-to-br from-purple-900 to-blue-900">
                  <div className="absolute -bottom-12 left-6">
                    <div className="w-24 h-24 rounded-full bg-gray-700 overflow-hidden border-4 border-gray-800">
                      {artist.profile_image_path ? (
                        <img
                          src={getFileUrl(artist.profile_image_path)}
                          alt={artist.artist_name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center text-3xl text-gray-400">
                          {artist.artist_name.charAt(0).toUpperCase()}
                        </div>
                      )}
                    </div>
                  </div>
                </div>

                {/* Artist Info */}
                <div className="pt-16 px-6 pb-4">
                  <h3 className="text-xl font-bold mb-1 truncate">{artist.artist_name}</h3>
                  {artist.bio && (
                    <p className="text-sm text-gray-400 mb-4 line-clamp-2">{artist.bio}</p>
                  )}

                  {/* Stats */}
                  {stats && (
                    <div className="grid grid-cols-3 gap-4 mb-4 py-4 border-t border-b border-gray-700">
                      <div className="text-center">
                        <div className="text-2xl font-bold">{stats.trackCount}</div>
                        <div className="text-xs text-gray-400">Tracks</div>
                      </div>
                      <div className="text-center">
                        <div className="text-2xl font-bold">{stats.albumCount}</div>
                        <div className="text-xs text-gray-400">Albums</div>
                      </div>
                      <div className="text-center">
                        <div className="text-2xl font-bold">{stats.followerCount}</div>
                        <div className="text-xs text-gray-400">Followers</div>
                      </div>
                    </div>
                  )}

                  {/* Quick Actions */}
                  <div className="space-y-2">
                    <Link
                      href={`/artists/${artist.uuid}/upload`}
                      className="flex items-center justify-center gap-2 w-full bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold text-sm"
                    >
                      <Upload className="w-4 h-4" />
                      Upload Music
                    </Link>
                    <div className="grid grid-cols-3 gap-2">
                      <Link
                        href={`/artists/${artist.uuid}/edit`}
                        className="flex items-center justify-center gap-1 bg-gray-700 hover:bg-gray-600 px-3 py-2 rounded-lg text-xs"
                        title="Edit Artist"
                      >
                        <Edit className="w-4 h-4" />
                      </Link>
                      <Link
                        href={`/artists/${artist.uuid}/members`}
                        className="flex items-center justify-center gap-1 bg-gray-700 hover:bg-gray-600 px-3 py-2 rounded-lg text-xs"
                        title="Manage Members"
                      >
                        <Users className="w-4 h-4" />
                      </Link>
                      <Link
                        href={`/artists/${artist.uuid}`}
                        className="flex items-center justify-center gap-1 bg-gray-700 hover:bg-gray-600 px-3 py-2 rounded-lg text-xs"
                        title="View Public Page"
                      >
                        <Eye className="w-4 h-4" />
                      </Link>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="bg-gray-800 rounded-lg p-12 text-center">
          <MusicIcon className="w-16 h-16 text-gray-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-3">No Artists Yet</h2>
          <p className="text-gray-400 mb-6">
            Create your first artist profile to start uploading and sharing your music
          </p>
          <Link
            href="/artists/create"
            className="inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-700 px-6 py-3 rounded-lg font-semibold"
          >
            <UserPlus className="w-5 h-5" />
            Create Your First Artist
          </Link>
        </div>
      )}
    </div>
  );
}
