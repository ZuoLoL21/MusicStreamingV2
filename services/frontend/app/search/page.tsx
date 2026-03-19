'use client';

import { useState } from 'react';
import { api, getFileUrl } from '@/lib/api';
import { formatDuration } from '@/lib/formatters';
import { Music, Artist, Album, User, Playlist } from '@/lib/types';
import { Search as SearchIcon, Play } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import Link from 'next/link';
import toast from 'react-hot-toast';
import { AddToPlaylistButton } from '@/components/AddToPlaylistButton';

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const [musicResults, setMusicResults] = useState<Music[]>([]);
  const [artistResults, setArtistResults] = useState<Artist[]>([]);
  const [albumResults, setAlbumResults] = useState<Album[]>([]);
  const [userResults, setUserResults] = useState<User[]>([]);
  const [playlistResults, setPlaylistResults] = useState<Playlist[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchPerformed, setSearchPerformed] = useState(false);
  const { playQueue } = usePlayerStore();

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    setSearchPerformed(true);

    try {
      const [music, artists, albums, users, playlists] = await Promise.all([
        api.searchMusic(query),
        api.searchArtists(query),
        api.searchAlbums(query),
        api.searchUsers(query),
        api.searchPlaylists(query),
      ]);

      setMusicResults(music);
      setArtistResults(artists);
      setAlbumResults(albums);
      setUserResults(users);
      setPlaylistResults(playlists);
    } catch (error) {
      toast.error('Search failed');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handlePlayMusic = (music: Music, index: number) => {
    playQueue(musicResults, index);
  };

  return (
    <div className="p-8">
      <h1 className="text-4xl font-bold mb-8">Search</h1>

      {/* Search Form */}
      <form onSubmit={handleSearch} className="mb-8">
        <div className="relative max-w-2xl">
          <SearchIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search for music, artists, or albums..."
            className="w-full pl-12 pr-4 py-3 bg-gray-900 border border-gray-700 rounded-full focus:outline-none focus:ring-2 focus:ring-green-500"
          />
        </div>
      </form>

      {loading && (
        <div className="text-center text-gray-400">Searching...</div>
      )}

      {!loading && searchPerformed && (
        <>
          {/* Artists Results */}
          {artistResults.length > 0 && (
            <section className="mb-12">
              <h2 className="text-2xl font-bold mb-4">Artists</h2>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
                {artistResults.map((artist) => (
                  <Link
                    key={artist.uuid}
                    href={`/artists/${artist.uuid}`}
                    className="bg-gray-900 p-4 rounded-lg hover:bg-gray-800 transition"
                  >
                    <div className="aspect-square bg-gray-800 rounded-full mb-4 flex items-center justify-center overflow-hidden">
                      {artist.profile_image_path ? (
                        <img
                          src={getFileUrl(artist.profile_image_path)}
                          alt={artist.artist_name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="text-4xl text-gray-600">
                          {artist.artist_name.charAt(0).toUpperCase()}
                        </div>
                      )}
                    </div>
                    <h3 className="font-semibold text-center truncate">{artist.artist_name}</h3>
                    <p className="text-sm text-gray-400 text-center">Artist</p>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Albums Results */}
          {albumResults.length > 0 && (
            <section className="mb-12">
              <h2 className="text-2xl font-bold mb-4">Albums</h2>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
                {albumResults.map((album) => (
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
                    <p className="text-sm text-gray-400 truncate">Album</p>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Users Results */}
          {userResults.length > 0 && (
            <section className="mb-12">
              <h2 className="text-2xl font-bold mb-4">Users</h2>
              <div className="space-y-2">
                {userResults.map((user) => (
                  <Link
                    key={user.uuid}
                    href={`/profile/${user.uuid}`}
                    className="flex items-center gap-4 p-4 bg-gray-900 rounded-lg hover:bg-gray-800 transition"
                  >
                    <img
                      src={getFileUrl(user.profile_image_path || '')}
                      alt={user.username}
                      className="w-12 h-12 rounded-full object-cover"
                    />
                    <div>
                      <h3 className="font-semibold">{user.username}</h3>
                      {user.bio && <p className="text-sm text-gray-400 line-clamp-1">{user.bio}</p>}
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Playlists Results */}
          {playlistResults.length > 0 && (
            <section className="mb-12">
              <h2 className="text-2xl font-bold mb-4">Playlists</h2>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
                {playlistResults.map((playlist) => (
                  <Link
                    key={playlist.uuid}
                    href={`/playlists/${playlist.uuid}`}
                    className="bg-gray-900 p-4 rounded-lg hover:bg-gray-800 transition"
                  >
                    <div className="aspect-square bg-gray-800 rounded mb-4 overflow-hidden">
                      {playlist.image_path ? (
                        <img
                          src={getFileUrl(playlist.image_path)}
                          alt={playlist.original_name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center text-gray-600">
                          <Play className="w-12 h-12" />
                        </div>
                      )}
                    </div>
                    <h3 className="font-semibold truncate">{playlist.original_name}</h3>
                    <p className="text-sm text-gray-400 truncate">
                      {playlist.is_public ? 'Public' : 'Private'} Playlist
                    </p>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Music Results */}
          {musicResults.length > 0 && (
            <section>
              <h2 className="text-2xl font-bold mb-4">Songs</h2>
              <div className="space-y-2">
                {musicResults.map((music, index) => (
                  <div
                    key={music.uuid}
                    className="flex items-center space-x-4 p-3 bg-gray-900 hover:bg-gray-800 rounded-lg group"
                  >
                    <div className="w-12 h-12 bg-gray-800 rounded overflow-hidden flex-shrink-0">
                      {music.image_path ? (
                        <img
                          src={getFileUrl(music.image_path)}
                          alt={music.song_name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center text-gray-600">
                          <Play className="w-6 h-6" />
                        </div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0 cursor-pointer" onClick={() => handlePlayMusic(music, index)}>
                      <h3 className="font-semibold truncate">{music.song_name}</h3>
                      <p className="text-sm text-gray-400">Artist Name</p>
                    </div>
                    <div className="text-sm text-gray-400">
                      {formatDuration(music.duration_seconds)}
                    </div>
                    <div className="opacity-0 group-hover:opacity-100 transition">
                      <AddToPlaylistButton musicUuid={music.uuid} size="sm" />
                    </div>
                    <button
                      onClick={() => handlePlayMusic(music, index)}
                      className="w-10 h-10 bg-green-500 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition"
                    >
                      <Play className="w-5 h-5 text-black ml-1" />
                    </button>
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* No Results */}
          {musicResults.length === 0 &&
            artistResults.length === 0 &&
            albumResults.length === 0 &&
            userResults.length === 0 &&
            playlistResults.length === 0 && (
              <div className="text-center text-gray-400 mt-8">
                No results found for "{query}"
              </div>
            )}
        </>
      )}
    </div>
  );
}
