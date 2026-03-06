'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api, getFileUrl } from '@/lib/api';
import { Artist, Music, Album } from '@/lib/types';
import { Play } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';

export default function ArtistPage() {
  const params = useParams();
  const artistId = params.id as string;
  const [artist, setArtist] = useState<Artist | null>(null);
  const [music, setMusic] = useState<Music[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState<'owner' | 'manager' | 'member' | null>(null);
  const { playQueue } = usePlayerStore();

  useEffect(() => {
    loadArtistData();
  }, [artistId]);

  const loadArtistData = async () => {
    try {
      setLoading(true);
      const [artistData, musicData, albumsData] = await Promise.all([
        api.getArtist(artistId),
        api.getArtistMusic(artistId, 50),
        api.getArtistAlbums(artistId, 20),
      ]);

      setArtist(artistData);
      setMusic(musicData);
      setAlbums(albumsData);

      // Check if current user is a member
      try {
        const currentUser = await api.getCurrentUser();
        const members = await api.getArtistMembers(artistId);
        const userMember = members.find((m) => m.user_uuid === currentUser.uuid);
        if (userMember) {
          setUserRole(userMember.role);
        }
      } catch (e) {
        // User is not logged in or not a member
      }
    } catch (error) {
      toast.error('Failed to load artist');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handlePlayAll = () => {
    if (music.length > 0) {
      playQueue(music, 0);
    }
  };

  const handlePlayMusic = (track: Music, index: number) => {
    playQueue(music, index);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Loading...</div>
      </div>
    );
  }

  if (!artist) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Artist not found</div>
      </div>
    );
  }

  return (
    <div>
      {/* Artist Header */}
      <div className="bg-gradient-to-b from-green-900 to-black p-8">
        <div className="flex items-end space-x-6">
          <div className="w-48 h-48 bg-gray-800 rounded-full overflow-hidden flex-shrink-0">
            {artist.profile_image_path ? (
              <img
                src={getFileUrl(artist.profile_image_path)}
                alt={artist.artist_name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-6xl text-gray-600">
                {artist.artist_name.charAt(0).toUpperCase()}
              </div>
            )}
          </div>
          <div className="flex-1">
            <p className="text-sm font-semibold uppercase">Artist</p>
            <h1 className="text-6xl font-bold my-2">{artist.artist_name}</h1>
            {artist.follower_count !== undefined && (
              <p className="text-gray-400">{artist.follower_count.toLocaleString()} followers</p>
            )}
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="bg-gradient-to-b from-black/60 to-black p-8">
        <div className="flex items-center gap-4">
          <button
            onClick={handlePlayAll}
            disabled={music.length === 0}
            className="w-14 h-14 bg-green-500 rounded-full flex items-center justify-center hover:scale-105 transition disabled:opacity-50"
          >
            <Play className="w-7 h-7 text-black ml-1" />
          </button>

          {/* Management Actions */}
          {userRole && (
            <div className="flex gap-3">
              {(userRole === 'owner' || userRole === 'manager') && (
                <>
                  <Link
                    href={`/artists/${artistId}/edit`}
                    className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-full font-semibold"
                  >
                    Edit Artist
                  </Link>
                  <Link
                    href={`/artists/${artistId}/upload`}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-full font-semibold"
                  >
                    Upload Music
                  </Link>
                  <Link
                    href={`/artists/${artistId}/albums/create`}
                    className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-full font-semibold"
                  >
                    Create Album
                  </Link>
                </>
              )}
              {userRole === 'member' && (
                <Link
                  href={`/artists/${artistId}/upload`}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-full font-semibold"
                >
                  Upload Music
                </Link>
              )}
              {userRole === 'owner' && (
                <Link
                  href={`/artists/${artistId}/members`}
                  className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-full font-semibold"
                >
                  Manage Members
                </Link>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="p-8">
        {/* Popular Tracks */}
        {music.length > 0 && (
          <section className="mb-12">
            <h2 className="text-2xl font-bold mb-4">Popular</h2>
            <div className="space-y-2">
              {music.slice(0, 10).map((track, index) => (
                <div
                  key={track.uuid}
                  onClick={() => handlePlayMusic(track, index)}
                  className="flex items-center space-x-4 p-3 hover:bg-gray-900 rounded-lg cursor-pointer group"
                >
                  <span className="text-gray-400 w-6 text-center">{index + 1}</span>
                  <div className="w-12 h-12 bg-gray-800 rounded overflow-hidden flex-shrink-0">
                    {track.image_path ? (
                      <img
                        src={getFileUrl(track.image_path)}
                        alt={track.song_name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-gray-600">
                        <Play className="w-6 h-6" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-semibold truncate">{track.song_name}</h3>
                    <p className="text-sm text-gray-400">{track.play_count.toLocaleString()} plays</p>
                  </div>
                  <div className="text-sm text-gray-400">
                    {Math.floor(track.duration_seconds / 60)}:{String(track.duration_seconds % 60).padStart(2, '0')}
                  </div>
                  <button className="w-10 h-10 bg-green-500 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition">
                    <Play className="w-5 h-5 text-black ml-1" />
                  </button>
                </div>
              ))}
            </div>
          </section>
        )}

        {/* Albums */}
        {albums.length > 0 && (
          <section>
            <h2 className="text-2xl font-bold mb-4">Albums</h2>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {albums.map((album) => (
                <Link
                  key={album.uuid}
                  href={`/albums/${album.uuid}`}
                  className="bg-gray-900 p-4 rounded-lg hover:bg-gray-800 transition"
                >
                  <div className="aspect-square bg-gray-800 rounded mb-4 overflow-hidden">
                    {album.image_path ? (
                      <img
                        src={getFileUrl(album.image_path)}
                        alt={album.original_name}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center text-gray-600">
                        <Play className="w-12 h-12" />
                      </div>
                    )}
                  </div>
                  <h3 className="font-semibold truncate">{album.original_name}</h3>
                  <p className="text-sm text-gray-400">Album</p>
                </Link>
              ))}
            </div>
          </section>
        )}
      </div>
    </div>
  );
}
