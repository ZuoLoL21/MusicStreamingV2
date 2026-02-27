import { create } from 'zustand';
import { Music } from './types';

interface PlayerState {
  currentTrack: Music | null;
  queue: Music[];
  currentIndex: number;
  isPlaying: boolean;
  playTrack: (track: Music) => void;
  playQueue: (tracks: Music[], startIndex?: number) => void;
  playNext: () => void;
  playPrevious: () => void;
  setIsPlaying: (playing: boolean) => void;
}

export const usePlayerStore = create<PlayerState>((set, get) => ({
  currentTrack: null,
  queue: [],
  currentIndex: 0,
  isPlaying: false,

  playTrack: (track) => {
    set({ currentTrack: track, queue: [track], currentIndex: 0, isPlaying: true });
  },

  playQueue: (tracks, startIndex = 0) => {
    set({
      queue: tracks,
      currentTrack: tracks[startIndex],
      currentIndex: startIndex,
      isPlaying: true,
    });
  },

  playNext: () => {
    const { queue, currentIndex } = get();
    if (currentIndex < queue.length - 1) {
      const newIndex = currentIndex + 1;
      set({
        currentTrack: queue[newIndex],
        currentIndex: newIndex,
        isPlaying: true,
      });
    }
  },

  playPrevious: () => {
    const { queue, currentIndex } = get();
    if (currentIndex > 0) {
      const newIndex = currentIndex - 1;
      set({
        currentTrack: queue[newIndex],
        currentIndex: newIndex,
        isPlaying: true,
      });
    }
  },

  setIsPlaying: (playing) => set({ isPlaying: playing }),
}));

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  userUuid: string | null;
  setAuth: (normalToken: string, refreshToken: string, userUuid: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  refreshToken: null,
  userUuid: null,
  setAuth: (normalToken, refreshToken, userUuid) => {
    set({ token: normalToken, refreshToken, userUuid });
  },
  clearAuth: () => {
    set({ token: null, refreshToken: null, userUuid: null });
  },
}));
