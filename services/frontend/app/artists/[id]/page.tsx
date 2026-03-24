'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { api, getFileUrl } from '@/lib/api';
import { formatDuration } from '@/lib/formatters';
import { Artist, Music, Album } from '@/lib/types';
import { Play, Upload, Edit, Users, BarChart3, UserPlus, UserMinus } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function ArtistPage() {
  const params = useParams();
  const artistId = params.id as string;
  const [artist, setArtist] = useState<Artist | null>(null);
  const [music, setMusic] = useState<Music[]>([]);
  const [albums, setAlbums] = useState<Album[]>([]);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState<'owner' | 'manager' | 'member' | null>(null);
  const [isFollowing, setIsFollowing] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
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

      // Check if current user is a member and if following
      try {
        const currentUser = await api.getCurrentUser();
        setIsLoggedIn(true);

        const members = await api.getArtistMembers(artistId);
        console.log('Current user UUID:', currentUser.uuid);
        console.log('Artist members:', members);
        const userMember = members.find((m) => m.uuid === currentUser.uuid);
        console.log('User member found:', userMember);
        if (userMember) {
          setUserRole(userMember.role);
        }

        // Check if following (only if not a member)
        if (!userMember) {
          try {
            const followedArtists = await api.getFollowedArtists(currentUser.uuid, 1000);
            const isFollowingArtist = followedArtists.some(a => a.uuid === artistId);
            setIsFollowing(isFollowingArtist);
          } catch (e) {
            // Couldn't check following status
          }
        }
      } catch (e) {
        // User is not logged in or not a member
        setIsLoggedIn(false);
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

  const handleFollowToggle = async () => {
    if (!isLoggedIn) {
      toast.error('Please log in to follow artists');
      return;
    }

    try {
      if (isFollowing) {
        await api.unfollowArtist(artistId);
        setIsFollowing(false);
        toast.success('Unfollowed artist');
        // Update follower count
        if (artist) {
          setArtist({
            ...artist,
            follower_count: (artist.follower_count || 0) - 1,
          });
        }
      } else {
        await api.followArtist(artistId);
        setIsFollowing(true);
        toast.success('Following artist');
        // Update follower count
        if (artist) {
          setArtist({
            ...artist,
            follower_count: (artist.follower_count || 0) + 1,
          });
        }
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update follow status');
    }
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

  // Management View for Owners/Managers
  const renderManagementView = () => (
    <div>
      {/* Management Header */}
      <div className="bg-gradient-to-b from-purple-900 to-black p-8">
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
            <div className="flex items-center gap-3 mb-2">
              <p className="text-sm font-semibold uppercase">Your Artist</p>
              <span className="px-2 py-1 bg-purple-600 text-xs rounded-full">{userRole}</span>
            </div>
            <h1 className="text-6xl font-bold my-2">{artist.artist_name}</h1>
            {artist.follower_count !== undefined && (
              <p className="text-gray-400">{artist.follower_count.toLocaleString()} followers</p>
            )}
          </div>
        </div>
      </div>

      {/* Management Stats & Actions */}
      <div className="bg-gradient-to-b from-black/60 to-black p-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 mb-6">
          {/* Stats Cards */}
          <div className="bg-gray-800 p-4 rounded-lg">
            <div className="text-3xl font-bold">{music.length}</div>
            <div className="text-sm text-gray-400">Total Tracks</div>
          </div>
          <div className="bg-gray-800 p-4 rounded-lg">
            <div className="text-3xl font-bold">{albums.length}</div>
            <div className="text-sm text-gray-400">Albums</div>
          </div>
          <div className="bg-gray-800 p-4 rounded-lg">
            <div className="text-3xl font-bold">{artist.follower_count || 0}</div>
            <div className="text-sm text-gray-400">Followers</div>
          </div>
          <div className="bg-gray-800 p-4 rounded-lg">
            <div className="text-3xl font-bold">
              {music.reduce((sum, m) => sum + m.play_count, 0).toLocaleString()}
            </div>
            <div className="text-sm text-gray-400">Total Plays</div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="flex items-center gap-3 flex-wrap">
          <button
            onClick={handlePlayAll}
            disabled={music.length === 0}
            className="w-14 h-14 bg-green-500 rounded-full flex items-center justify-center hover:scale-105 transition disabled:opacity-50"
          >
            <Play className="w-7 h-7 text-black ml-1" />
          </button>

          <Link
            href={`/artists/${artistId}/upload`}
            className="flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 rounded-full font-semibold"
          >
            <Upload className="w-5 h-5" />
            Upload Music
          </Link>

          {(userRole === 'owner' || userRole === 'manager') && (
            <>
              <Link
                href={`/artists/${artistId}/edit`}
                className="flex items-center gap-2 px-4 py-3 bg-gray-700 hover:bg-gray-600 rounded-full font-semibold"
              >
                <Edit className="w-5 h-5" />
                Edit Artist
              </Link>
              <Link
                href={`/artists/${artistId}/albums/create`}
                className="flex items-center gap-2 px-4 py-3 bg-gray-700 hover:bg-gray-600 rounded-full font-semibold"
              >
                <Upload className="w-5 h-5" />
                Create Album
              </Link>
            </>
          )}

          {userRole === 'owner' && (
            <Link
              href={`/artists/${artistId}/members`}
              className="flex items-center gap-2 px-4 py-3 bg-gray-700 hover:bg-gray-600 rounded-full font-semibold"
            >
              <Users className="w-5 h-5" />
              Manage Members
            </Link>
          )}
        </div>
      </div>
    </div>
  );

  // Listener View for Regular Users
  const renderListenerView = () => (
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
            {artist.bio && <p className="text-gray-400 mt-2">{artist.bio}</p>}
            {artist.follower_count !== undefined && (
              <p className="text-gray-400 mt-2">{artist.follower_count.toLocaleString()} followers</p>
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

          {isLoggedIn && (
            <button
              onClick={handleFollowToggle}
              className={`flex items-center gap-2 px-6 py-3 rounded-full font-semibold ${
                isFollowing
                  ? 'bg-gray-700 hover:bg-gray-600'
                  : 'bg-blue-600 hover:bg-blue-700'
              }`}
            >
              {isFollowing ? (
                <>
                  <UserMinus className="w-5 h-5" />
                  Unfollow
                </>
              ) : (
                <>
                  <UserPlus className="w-5 h-5" />
                  Follow
                </>
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  );

  return (
    <div>
      {/* Render appropriate view based on user role */}
      {userRole ? renderManagementView() : renderListenerView()}

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
                  className="flex items-center space-x-4 p-3 hover:bg-gray-900 rounded-lg group"
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
                  <div className="flex-1 min-w-0 cursor-pointer" onClick={() => handlePlayMusic(track, index)}>
                    <h3 className="font-semibold truncate">{track.song_name}</h3>
                    <p className="text-sm text-gray-400">{track.play_count.toLocaleString()} plays</p>
                  </div>
                  <div className="text-sm text-gray-400">
                    {formatDuration(track.duration_seconds)}
                  </div>
                  <div className="opacity-0 group-hover:opacity-100 transition">
                    <AddToPlaylistButton musicUuid={track.uuid} size="sm" />
                  </div>
                  <button
                    onClick={() => handlePlayMusic(track, index)}
                    className="w-10 h-10 bg-green-500 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition"
                  >
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
