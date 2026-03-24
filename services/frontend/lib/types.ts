export interface User {
  uuid: string;
  username: string;
  email: string;
  bio?: string;
  profile_image_path?: string;
  country: string;
  created_at: string;
  similarity_score?: number; // Only present in search results
}

export interface Artist {
  uuid: string;
  artist_name: string;
  bio?: string;
  profile_image_path?: string;
  follower_count?: number;
  created_at: string;
  similarity_score?: number; // Only present in search results
}

export interface Album {
  uuid: string;
  from_artist: string;
  original_name: string;
  description?: string;
  image_path?: string;
  created_at: string;
  similarity_score?: number; // Only present in search results
}

export interface Music {
  uuid: string;
  from_artist: string;
  uploaded_by: string;
  in_album?: string;
  song_name: string;
  path_in_file_storage: string;
  image_path?: string;
  play_count: number;
  duration_seconds: number;
  created_at: string;
  similarity_score?: number; // Only present in search results
}

export interface Playlist {
  uuid: string;
  from_user: string;
  original_name: string;
  description?: string;
  is_public: boolean;
  image_path?: string;
  created_at: string;
  similarity_score?: number; // Only present in search results
}

export interface Tag {
  tag_name: string;
  tag_description?: string;
  created_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user_uuid: string;
  device_id: string;
}

// Cursor types for pagination
export interface Cursor {
  cursor_ts?: string;
  cursor_id?: string;
}

export interface SearchCursor {
  cursor_score?: number;
  cursor_ts?: string;
}

export interface PopularityCursor {
  cursor_decay?: number;
  cursor_plays?: number;
  cursor_id?: string;
}

// Artist Member Management
export interface ArtistMember {
  user_uuid: string;
  username: string;
  role: 'owner' | 'manager' | 'member';
  added_at: string;
}

// Recommendations & Discovery
export interface ThemeRecommendation {
  recommended_theme: string;
  theme_features: number[];
  popularity_data: ThemePopularity[];
}

export interface SongPopularity {
  music_uuid: string;
  artist_uuid?: string;
  song_name: string;
  artist_name: string;
  decay_plays?: number;
  plays?: number;
  decay_listen_seconds?: number;
  listen_seconds?: number;
}

export interface ArtistPopularity {
  artist_uuid: string;
  artist_name: string;
  profile_image_path?: string;
  decay_plays?: number;
  plays?: number;
  decay_listen_seconds?: number;
  listen_seconds?: number;
}

export interface ThemePopularity {
  theme: string;
  decay_plays: number;
  decay_listen_seconds: number;
}

// History & Analytics
export interface ListeningHistory {
  music_uuid: string;
  song_name: string;
  artist_name: string;
  artist_uuid: string;
  listened_at: string;
  listen_duration_seconds?: number;
  completion_percentage?: number;
}

export interface TopMusic {
  music_uuid: string;
  song_name: string;
  artist_name: string;
  artist_uuid: string;
  play_count: number;
}

// Playlist Track
export interface PlaylistTrack {
  track_uuid: string;
  music_uuid: string;
  position: number;
  added_at: string;
  // Music fields
  song_name?: string;
  from_artist?: string;
  duration_seconds?: number;
  image_path?: string;
}

// Device Management
export interface Device {
  uuid: string;
  device_id: string;
  device_name: string | null;
  created_at: string;
  last_used_at: string;
  expires_at: string;
}