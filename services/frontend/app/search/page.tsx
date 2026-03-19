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

interface SearchCursor {
  cursor_score?: number;
  cursor_ts?: string;
}

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const [musicResults, setMusicResults] = useState<Music[]>([]);
  const [artistResults, setArtistResults] = useState<Artist[]>([]);
  const [albumResults, setAlbumResults] = useState<Album[]>([]);
  const [userResults, setUserResults] = useState<User[]>([]);
  const [playlistResults, setPlaylistResults] = useState<Playlist[]>([]);

  const [musicCursor, setMusicCursor] = useState<SearchCursor | undefined>();
  const [artistCursor, setArtistCursor] = useState<SearchCursor | undefined>();
  const [albumCursor, setAlbumCursor] = useState<SearchCursor | undefined>();
  const [userCursor, setUserCursor] = useState<SearchCursor | undefined>();
  const [playlistCursor, setPlaylistCursor] = useState<SearchCursor | undefined>();

  const [hasMoreMusic, setHasMoreMusic] = useState(false);
  const [hasMoreArtists, setHasMoreArtists] = useState(false);
  const [hasMoreAlbums, setHasMoreAlbums] = useState(false);
  const [hasMoreUsers, setHasMoreUsers] = useState(false);
  const [hasMorePlaylists, setHasMorePlaylists] = useState(false);

  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState<string | null>(null);
  const [searchPerformed, setSearchPerformed] = useState(false);
  const { playQueue } = usePlayerStore();

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;

    setLoading(true);
    setSearchPerformed(true);
    // Reset cursors on new search
    setMusicCursor(undefined);
    setArtistCursor(undefined);
    setAlbumCursor(undefined);
    setUserCursor(undefined);
    setPlaylistCursor(undefined);

    try {
      const [musicRes, artistsRes, albumsRes, usersRes, playlistsRes] =
        await Promise.all([
          api.searchMusic(query),
          api.searchArtists(query),
          api.searchAlbums(query),
          api.searchUsers(query),
          api.searchPlaylists(query),
        ]);

      setMusicResults(musicRes.music);
      setHasMoreMusic(musicRes.hasMore);
      if (musicRes.music.length > 0) {
        const lastItem = musicRes.music[musicRes.music.length - 1];
        setMusicCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }

      setArtistResults(artistsRes.artists);
      setHasMoreArtists(artistsRes.hasMore);
      if (artistsRes.artists.length > 0) {
        const lastItem = artistsRes.artists[artistsRes.artists.length - 1];
        setArtistCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }

      setAlbumResults(albumsRes.albums);
      setHasMoreAlbums(albumsRes.hasMore);
      if (albumsRes.albums.length > 0) {
        const lastItem = albumsRes.albums[albumsRes.albums.length - 1];
        setAlbumCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }

      setUserResults(usersRes.users);
      setHasMoreUsers(usersRes.hasMore);
      if (usersRes.users.length > 0) {
        const lastItem = usersRes.users[usersRes.users.length - 1];
        setUserCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }

      setPlaylistResults(playlistsRes.playlists);
      setHasMorePlaylists(playlistsRes.hasMore);
      if (playlistsRes.playlists.length > 0) {
        const lastItem = playlistsRes.playlists[playlistsRes.playlists.length - 1];
        setPlaylistCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }
    } catch (error) {
      toast.error('Search failed');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const loadMoreMusic = async () => {
    if (!hasMoreMusic || !musicCursor || loadingMore) return;
    setLoadingMore('music');
    try {
      const musicRes = await api.searchMusic(query, 20, musicCursor);
      setMusicResults([...musicResults, ...musicRes.music]);
      setHasMoreMusic(musicRes.hasMore);

      if (musicRes.music.length > 0) {
        const lastItem = musicRes.music[musicRes.music.length - 1];
        setMusicCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }
    } catch (error) {
      toast.error('Failed to load more songs');
    } finally {
      setLoadingMore(null);
    }
  };

  const loadMoreArtists = async () => {
    if (!hasMoreArtists || !artistCursor || loadingMore) return;
    setLoadingMore('artists');
    try {
      const artistsRes = await api.searchArtists(query, 20, artistCursor);
      setArtistResults([...artistResults, ...artistsRes.artists]);
      setHasMoreArtists(artistsRes.hasMore);

      if (artistsRes.artists.length > 0) {
        const lastItem = artistsRes.artists[artistsRes.artists.length - 1];
        setArtistCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }
    } catch (error) {
      toast.error('Failed to load more artists');
    } finally {
      setLoadingMore(null);
    }
  };

  const loadMoreAlbums = async () => {
    if (!hasMoreAlbums || !albumCursor || loadingMore) return;
    setLoadingMore('albums');
    try {
      const albumsRes = await api.searchAlbums(query, 20, albumCursor);
      setAlbumResults([...albumResults, ...albumsRes.albums]);
      setHasMoreAlbums(albumsRes.hasMore);

      if (albumsRes.albums.length > 0) {
        const lastItem = albumsRes.albums[albumsRes.albums.length - 1];
        setAlbumCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }
    } catch (error) {
      toast.error('Failed to load more albums');
    } finally {
      setLoadingMore(null);
    }
  };

  const loadMoreUsers = async () => {
    if (!hasMoreUsers || !userCursor || loadingMore) return;
    setLoadingMore('users');
    try {
      const usersRes = await api.searchUsers(query, 20, userCursor);
      setUserResults([...userResults, ...usersRes.users]);
      setHasMoreUsers(usersRes.hasMore);

      if (usersRes.users.length > 0) {
        const lastItem = usersRes.users[usersRes.users.length - 1];
        setUserCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }
    } catch (error) {
      toast.error('Failed to load more users');
    } finally {
      setLoadingMore(null);
    }
  };

  const loadMorePlaylists = async () => {
    if (!hasMorePlaylists || !playlistCursor || loadingMore) return;
    setLoadingMore('playlists');
    try {
      const playlistsRes = await api.searchPlaylists(query, 20, playlistCursor);
      setPlaylistResults([...playlistResults, ...playlistsRes.playlists]);
      setHasMorePlaylists(playlistsRes.hasMore);

      if (playlistsRes.playlists.length > 0) {
        const lastItem = playlistsRes.playlists[playlistsRes.playlists.length - 1];
        setPlaylistCursor({
          cursor_score: lastItem.similarity_score,
          cursor_ts: lastItem.created_at
        });
      }
    } catch (error) {
      toast.error('Failed to load more playlists');
    } finally {
      setLoadingMore(null);
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
              {hasMoreArtists && (
                <div className="text-center mt-4">
                  <button
                    onClick={loadMoreArtists}
                    className="px-6 py-2 bg-green-500 text-white rounded-full hover:bg-green-600 transition disabled:opacity-50"
                    disabled={loadingMore === 'artists'}
                  >
                    {loadingMore === 'artists' ? 'Loading...' : 'Load More Artists'}
                  </button>
                </div>
              )}
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
              {hasMoreAlbums && (
                <div className="text-center mt-4">
                  <button
                    onClick={loadMoreAlbums}
                    className="px-6 py-2 bg-green-500 text-white rounded-full hover:bg-green-600 transition disabled:opacity-50"
                    disabled={loadingMore === 'albums'}
                  >
                    {loadingMore === 'albums' ? 'Loading...' : 'Load More Albums'}
                  </button>
                </div>
              )}
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
              {hasMoreUsers && (
                <div className="text-center mt-4">
                  <button
                    onClick={loadMoreUsers}
                    className="px-6 py-2 bg-green-500 text-white rounded-full hover:bg-green-600 transition disabled:opacity-50"
                    disabled={loadingMore === 'users'}
                  >
                    {loadingMore === 'users' ? 'Loading...' : 'Load More Users'}
                  </button>
                </div>
              )}
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
              {hasMorePlaylists && (
                <div className="text-center mt-4">
                  <button
                    onClick={loadMorePlaylists}
                    className="px-6 py-2 bg-green-500 text-white rounded-full hover:bg-green-600 transition disabled:opacity-50"
                    disabled={loadingMore === 'playlists'}
                  >
                    {loadingMore === 'playlists' ? 'Loading...' : 'Load More Playlists'}
                  </button>
                </div>
              )}
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
              {hasMoreMusic && (
                <div className="text-center mt-4">
                  <button
                    onClick={loadMoreMusic}
                    className="px-6 py-2 bg-green-500 text-white rounded-full hover:bg-green-600 transition disabled:opacity-50"
                    disabled={loadingMore === 'music'}
                  >
                    {loadingMore === 'music' ? 'Loading...' : 'Load More Songs'}
                  </button>
                </div>
              )}
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
