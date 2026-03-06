import axios, { AxiosInstance } from 'axios';
import Cookies from 'js-cookie';
import {
  User,
  Artist,
  Album,
  Music,
  Playlist,
  Tag,
  AuthResponse,
  Cursor,
  SearchCursor,
  PopularityCursor,
  ArtistMember,
  ThemeRecommendation,
  SongPopularity,
  ArtistPopularity,
  ThemePopularity,
  ListeningHistory,
  TopMusic,
  PlaylistTrack,
} from './types';

// Use different URLs for server-side and client-side
const API_BASE_URL = typeof window === 'undefined'
  ? (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080')
  : (process.env.NEXT_PUBLIC_API_URL_BROWSER || 'http://localhost:8080');

// Helper to convert relative file paths to full URLs
export const getFileUrl = (path: string): string => {
  if (!path) return '';
  if (path.startsWith('http://') || path.startsWith('https://')) return path;
  // Direct URL to backend - browser loads files from gateway API
  const baseUrl = 'http://localhost:8080';
  return `${baseUrl}${path.startsWith('/') ? '' : '/'}${path}`;
};

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      // Don't set default Content-Type - let axios handle it based on request data
    });

    // Add auth token to requests
    this.client.interceptors.request.use((config) => {
      const token = Cookies.get('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }

      // Set Content-Type to JSON only for non-FormData requests
      if (!(config.data instanceof FormData)) {
        config.headers['Content-Type'] = 'application/json';
      }

      return config;
    });
  }

  // Auth
  async login(email: string, password: string): Promise<AuthResponse> {
    const response = await this.client.post('/login', { email, password });
    return response.data;
  }

  async register(email: string, password: string, username: string, displayName: string, country: string): Promise<AuthResponse> {
    const formData = new FormData();
    formData.append('email', email);
    formData.append('password', password);
    formData.append('username', username);
    formData.append('display_name', displayName);
    formData.append('country', country);

    // Don't set Content-Type manually - axios will set it with the correct boundary
    const response = await this.client.put('/login', formData);
    return response.data;
  }

  // Users
  async getCurrentUser(): Promise<User> {
    const response = await this.client.get('/users/me');
    return response.data;
  }

  async getUser(uuid: string): Promise<User> {
    const response = await this.client.get(`/users/${uuid}`);
    return response.data;
  }

  async updateProfile(username: string, bio?: string): Promise<void> {
    await this.client.post('/users/me', { username, bio });
  }

  async updateEmail(oldPassword: string, email: string): Promise<void> {
    await this.client.post('/users/me/email', { old_password: oldPassword, email });
  }

  async updatePassword(oldPassword: string, newPassword: string): Promise<void> {
    await this.client.post('/users/me/password', { old_password: oldPassword, new_password: newPassword });
  }

  async uploadProfileImage(file: File): Promise<void> {
    const formData = new FormData();
    formData.append('image', file);
    await this.client.post('/users/me/image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  // Social - Follow/Unfollow
  async followUser(uuid: string): Promise<void> {
    await this.client.post(`/users/${uuid}/follow`);
  }

  async unfollowUser(uuid: string): Promise<void> {
    await this.client.delete(`/users/${uuid}/follow`);
  }

  async checkIfFollowingUser(uuid: string): Promise<{ is_following: boolean }> {
    const response = await this.client.get(`/users/${uuid}/following/check`);
    return response.data;
  }

  async followArtist(uuid: string): Promise<void> {
    await this.client.post(`/artists/${uuid}/follow`);
  }

  async unfollowArtist(uuid: string): Promise<void> {
    await this.client.delete(`/artists/${uuid}/follow`);
  }

  // Social - Followers & Following
  async getFollowersForUser(uuid: string, limit = 20, cursor?: Cursor): Promise<User[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/users/${uuid}/followers?${params}`);
    return response.data;
  }

  async getFollowingUsers(uuid: string, limit = 20, cursor?: Cursor): Promise<User[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/users/${uuid}/following/users?${params}`);
    return response.data;
  }

  async getFollowedArtists(uuid: string, limit = 20, cursor?: Cursor): Promise<Artist[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/users/${uuid}/following/artists?${params}`);
    return response.data;
  }

  async getLikedSongs(uuid: string, limit = 20, cursor?: Cursor): Promise<Music[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/users/${uuid}/likes?${params}`);
    return response.data;
  }

  async getUserPlaylists(uuid: string, limit = 20, cursor?: Cursor): Promise<Playlist[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/users/${uuid}/playlists?${params}`);
    return response.data;
  }

  async getUserArtists(uuid: string): Promise<Artist[]> {
    const response = await this.client.get(`/users/${uuid}/artists`);
    return response.data;
  }

  // Artists
  async getArtists(limit = 20, cursor?: string): Promise<Artist[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor) params.append('cursor_name', cursor);
    const response = await this.client.get(`/artists?${params}`);
    return response.data;
  }

  async getArtist(uuid: string): Promise<Artist> {
    const response = await this.client.get(`/artists/${uuid}`);
    return response.data;
  }

  async getArtistMusic(uuid: string, limit = 20): Promise<Music[]> {
    const response = await this.client.get(`/artists/${uuid}/music?limit=${limit}`);
    return response.data;
  }

  async getArtistAlbums(uuid: string, limit = 20): Promise<Album[]> {
    const response = await this.client.get(`/artists/${uuid}/albums?limit=${limit}`);
    return response.data;
  }

  async createArtist(name: string, bio?: string, image?: File): Promise<void> {
    const formData = new FormData();
    formData.append('artist_name', name);
    if (bio) formData.append('bio', bio);
    if (image) formData.append('image', image);
    await this.client.put('/artists', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async updateArtistProfile(uuid: string, name: string, bio?: string): Promise<void> {
    await this.client.post(`/artists/${uuid}`, { artist_name: name, bio });
  }

  async uploadArtistImage(uuid: string, image: File): Promise<void> {
    const formData = new FormData();
    formData.append('image', image);
    await this.client.post(`/artists/${uuid}/image`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async getArtistMembers(uuid: string): Promise<ArtistMember[]> {
    const response = await this.client.get(`/artists/${uuid}/members`);
    return response.data;
  }

  async addMemberToArtist(artistUuid: string, userUuid: string, role: string): Promise<void> {
    await this.client.put(`/artists/${artistUuid}/members/${userUuid}`, { role });
  }

  async removeMemberFromArtist(artistUuid: string, userUuid: string): Promise<void> {
    await this.client.delete(`/artists/${artistUuid}/members/${userUuid}`);
  }

  async changeArtistMemberRole(artistUuid: string, userUuid: string, role: string): Promise<void> {
    await this.client.post(`/artists/${artistUuid}/members/${userUuid}/role`, { role });
  }

  async getArtistFollowers(artistUuid: string, limit = 20, cursor?: Cursor): Promise<User[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/artists/${artistUuid}/followers?${params}`);
    return response.data;
  }

  // Albums
  async getAlbum(uuid: string): Promise<Album> {
    const response = await this.client.get(`/albums/${uuid}`);
    return response.data;
  }

  async getAlbumMusic(uuid: string): Promise<Music[]> {
    const response = await this.client.get(`/albums/${uuid}/music`);
    return response.data;
  }

  async createAlbum(artistUuid: string, name: string, description?: string, image?: File): Promise<void> {
    const formData = new FormData();
    formData.append('from_artist', artistUuid);
    formData.append('original_name', name);
    if (description) formData.append('description', description);
    if (image) formData.append('image', image);
    await this.client.put('/albums', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async updateAlbum(uuid: string, name: string, description?: string): Promise<void> {
    await this.client.post(`/albums/${uuid}`, { original_name: name, description });
  }

  async uploadAlbumImage(uuid: string, image: File): Promise<void> {
    const formData = new FormData();
    formData.append('image', image);
    await this.client.post(`/albums/${uuid}/image`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async deleteAlbum(uuid: string): Promise<void> {
    await this.client.delete(`/albums/${uuid}`);
  }

  // Music
  async getMusic(uuid: string): Promise<Music> {
    const response = await this.client.get(`/music/${uuid}`);
    return response.data;
  }

  async likeMusic(uuid: string): Promise<void> {
    await this.client.post(`/music/${uuid}/like`);
  }

  async unlikeMusic(uuid: string): Promise<void> {
    await this.client.delete(`/music/${uuid}/like`);
  }

  async checkIfMusicLiked(uuid: string): Promise<{ liked: boolean }> {
    const response = await this.client.get(`/music/${uuid}/liked`);
    return response.data;
  }

  async uploadMusic(
    artistUuid: string,
    songName: string,
    durationSeconds: number,
    audioFile: File,
    albumUuid?: string
  ): Promise<void> {
    const formData = new FormData();
    formData.append('from_artist', artistUuid);
    formData.append('song_name', songName);
    formData.append('duration_seconds', String(durationSeconds));
    formData.append('audio_file', audioFile);
    if (albumUuid) formData.append('in_album', albumUuid);
    await this.client.put('/music', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async updateMusicDetails(uuid: string, songName: string, albumUuid?: string): Promise<void> {
    await this.client.post(`/music/${uuid}`, {
      song_name: songName,
      in_album: albumUuid || null,
    });
  }

  async updateMusicStorage(uuid: string, audioFile: File, durationSeconds: number): Promise<void> {
    const formData = new FormData();
    formData.append('audio_file', audioFile);
    formData.append('duration_seconds', String(durationSeconds));
    await this.client.post(`/music/${uuid}/storage`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async deleteMusic(uuid: string): Promise<void> {
    await this.client.delete(`/music/${uuid}`);
  }

  async incrementPlayCount(uuid: string): Promise<void> {
    await this.client.post(`/music/${uuid}/play`);
  }

  async recordListeningHistory(
    uuid: string,
    listenDuration?: number,
    completionPercentage?: number
  ): Promise<void> {
    await this.client.post(`/music/${uuid}/listen`, {
      listen_duration_seconds: listenDuration,
      completion_percentage: completionPercentage,
    });
  }

  async getMusicTags(uuid: string, limit = 20, cursor?: string): Promise<Tag[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor) params.append('cursor_name', cursor);
    const response = await this.client.get(`/music/${uuid}/tags?${params}`);
    return response.data;
  }

  async assignTagToMusic(musicUuid: string, tagName: string): Promise<void> {
    await this.client.post(`/music/${musicUuid}/tags/${tagName}`);
  }

  async removeTagFromMusic(musicUuid: string, tagName: string): Promise<void> {
    await this.client.delete(`/music/${musicUuid}/tags/${tagName}`);
  }

  // Playlists
  async getPlaylists(limit = 20): Promise<Playlist[]> {
    const user = await this.getCurrentUser();
    return this.getUserPlaylists(user.uuid, limit);
  }

  async getPlaylist(uuid: string): Promise<Playlist> {
    const response = await this.client.get(`/playlists/${uuid}`);
    return response.data;
  }

  async getPlaylistTracks(uuid: string): Promise<Music[]> {
    const response = await this.client.get(`/playlists/${uuid}/tracks`);
    return response.data;
  }

  async createPlaylist(
    name: string,
    description?: string,
    isPublic?: boolean,
    image?: File
  ): Promise<void> {
    const formData = new FormData();
    formData.append('original_name', name);
    if (description) formData.append('description', description);
    formData.append('is_public', String(isPublic ?? true));
    if (image) formData.append('image', image);
    await this.client.put('/playlists', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async updatePlaylist(
    uuid: string,
    name: string,
    description?: string,
    isPublic?: boolean
  ): Promise<void> {
    await this.client.post(`/playlists/${uuid}`, {
      original_name: name,
      description,
      is_public: isPublic,
    });
  }

  async uploadPlaylistImage(uuid: string, image: File): Promise<void> {
    const formData = new FormData();
    formData.append('image', image);
    await this.client.post(`/playlists/${uuid}/image`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  }

  async deletePlaylist(uuid: string): Promise<void> {
    await this.client.delete(`/playlists/${uuid}`);
  }

  async addTrackToPlaylist(playlistUuid: string, musicUuid: string, position: number): Promise<void> {
    await this.client.put(`/playlists/${playlistUuid}/tracks/${musicUuid}`, { position });
  }

  async removeTrackFromPlaylist(playlistUuid: string, musicUuid: string): Promise<void> {
    await this.client.delete(`/playlists/${playlistUuid}/tracks/${musicUuid}`);
  }

  async updateTrackPosition(playlistUuid: string, trackUuid: string, position: number): Promise<void> {
    await this.client.post(`/playlists/${playlistUuid}/tracks/${trackUuid}/position`, { position });
  }

  // Tags
  async getTags(limit = 50): Promise<Tag[]> {
    const response = await this.client.get(`/tags?limit=${limit}`);
    return response.data;
  }

  async getTag(name: string): Promise<Tag> {
    const response = await this.client.get(`/tags/${name}`);
    return response.data;
  }

  async getMusicForTag(name: string, limit = 20): Promise<Music[]> {
    const response = await this.client.get(`/tags/${name}/music?limit=${limit}`);
    return response.data;
  }

  // Search
  async searchMusic(query: string, limit = 20): Promise<Music[]> {
    const response = await this.client.get(`/search/music?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data;
  }

  async searchArtists(query: string, limit = 20): Promise<Artist[]> {
    const response = await this.client.get(`/search/artists?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data;
  }

  async searchAlbums(query: string, limit = 20): Promise<Album[]> {
    const response = await this.client.get(`/search/albums?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data;
  }

  async searchUsers(query: string, limit = 20): Promise<User[]> {
    const response = await this.client.get(`/search/users?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data;
  }

  async searchPlaylists(query: string, limit = 20): Promise<Playlist[]> {
    const response = await this.client.get(`/search/playlists?q=${encodeURIComponent(query)}&limit=${limit}`);
    return response.data;
  }

  // Recommendations
  async getThemeRecommendation(): Promise<ThemeRecommendation> {
    const response = await this.client.post('/recommendation/theme');
    return response.data;
  }

  async getPopularSongsAllTime(limit = 20, cursor?: PopularityCursor): Promise<SongPopularity[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_decay) params.append('cursor_decay', String(cursor.cursor_decay));
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/recommendation/popular/songs/all-time?${params}`);
    return response.data;
  }

  async getPopularSongsTimeframe(
    startDate: string,
    endDate: string,
    limit = 20,
    cursor?: PopularityCursor
  ): Promise<SongPopularity[]> {
    const params = new URLSearchParams({
      limit: String(limit),
      start_date: startDate,
      end_date: endDate,
    });
    if (cursor?.cursor_plays) params.append('cursor_plays', String(cursor.cursor_plays));
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/recommendation/popular/songs/timeframe?${params}`);
    return response.data;
  }

  async getPopularSongsByTheme(
    theme: string,
    limit = 20,
    cursor?: PopularityCursor
  ): Promise<SongPopularity[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_plays) params.append('cursor_plays', String(cursor.cursor_plays));
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/recommendation/popular/songs/theme/${encodeURIComponent(theme)}?${params}`);
    return response.data;
  }

  async getPopularArtistsAllTime(limit = 20, cursor?: PopularityCursor): Promise<ArtistPopularity[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_decay) params.append('cursor_decay', String(cursor.cursor_decay));
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/recommendation/popular/artists/all-time?${params}`);
    return response.data;
  }

  async getPopularArtistsTimeframe(
    startDate: string,
    endDate: string,
    limit = 20,
    cursor?: PopularityCursor
  ): Promise<ArtistPopularity[]> {
    const params = new URLSearchParams({
      limit: String(limit),
      start_date: startDate,
      end_date: endDate,
    });
    if (cursor?.cursor_plays) params.append('cursor_plays', String(cursor.cursor_plays));
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/recommendation/popular/artists/timeframe?${params}`);
    return response.data;
  }

  async getPopularThemesAllTime(limit = 20): Promise<ThemePopularity[]> {
    const response = await this.client.get(`/recommendation/popular/themes/all-time?limit=${limit}`);
    return response.data;
  }

  async getPopularThemesTimeframe(
    startDate: string,
    endDate: string,
    limit = 20
  ): Promise<ThemePopularity[]> {
    const params = new URLSearchParams({
      limit: String(limit),
      start_date: startDate,
      end_date: endDate,
    });
    const response = await this.client.get(`/recommendation/popular/themes/timeframe?${params}`);
    return response.data;
  }

  // History & Analytics
  async getListeningHistory(limit = 20, cursor?: Cursor): Promise<ListeningHistory[]> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (cursor?.cursor_ts) params.append('cursor_ts', cursor.cursor_ts);
    if (cursor?.cursor_id) params.append('cursor_id', cursor.cursor_id);
    const response = await this.client.get(`/history?${params}`);
    return response.data;
  }

  async getTopMusicForUser(limit = 20): Promise<TopMusic[]> {
    const response = await this.client.get(`/history/top?limit=${limit}`);
    return response.data;
  }

  // Event Tracking
  async sendListenEvent(
    musicUuid: string,
    artistUuid: string,
    albumUuid: string | null,
    listenDurationSeconds: number,
    trackDurationSeconds: number
  ): Promise<void> {
    try {
      const completionRatio = trackDurationSeconds > 0 ? listenDurationSeconds / trackDurationSeconds : 0;

      await this.client.post('/events/listen', {
        music_uuid: musicUuid,
        artist_uuid: artistUuid,
        album_uuid: albumUuid,
        listen_duration_seconds: listenDurationSeconds,
        track_duration_seconds: trackDurationSeconds,
        completion_ratio: completionRatio,
      });
    } catch (error) {
      // Don't block user experience if event tracking fails
      console.warn('Failed to send listen event:', error);
    }
  }

  async sendLikeEvent(musicUuid: string, artistUuid: string): Promise<void> {
    try {
      await this.client.post('/events/like', {
        music_uuid: musicUuid,
        artist_uuid: artistUuid,
      });
    } catch (error) {
      // Don't block user experience if event tracking fails
      console.warn('Failed to send like event:', error);
    }
  }
}

export const api = new ApiClient();