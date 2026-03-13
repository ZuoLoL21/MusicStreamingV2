'use client';

import { useEffect, useRef, useState } from 'react';
import { Play, Pause, SkipBack, SkipForward, Volume2, Heart } from 'lucide-react';
import { usePlayerStore } from '@/lib/store';
import { api, getFileUrl } from '@/lib/api';
import Link from 'next/link';
import toast from 'react-hot-toast';

export function Player() {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [isLiked, setIsLiked] = useState(false);
  const startTimeRef = useRef<number>(0);
  const trackStartRef = useRef<number>(0);

  const { currentTrack, queue, currentIndex, playNext, playPrevious } = usePlayerStore();

  useEffect(() => {
    if (audioRef.current && currentTrack) {
      // Send listen event for previous track if it was playing
      if (startTimeRef.current > 0) {
        sendListenEvent();
      }

      audioRef.current.src = getFileUrl(currentTrack.path_in_file_storage);
      audioRef.current.play();
      setIsPlaying(true);

      // Track start time for listen event
      startTimeRef.current = Date.now();
      trackStartRef.current = 0;

      // Check if song is liked
      checkLiked();

      // Increment play count
      api.incrementPlayCount(currentTrack.uuid).catch((err) => console.error('Failed to increment play count:', err));
    }
  }, [currentTrack]);

  const sendListenEvent = () => {
    if (!currentTrack || startTimeRef.current === 0) return;

    const listenDuration = Math.floor((Date.now() - startTimeRef.current) / 1000);
    const trackDuration = currentTrack.duration_seconds;

    if (listenDuration > 0) {
      const completionPercentage = Math.round((listenDuration / trackDuration) * 100);
      api.recordListeningHistory(
        currentTrack.uuid,
        listenDuration,
        completionPercentage
      );
    }

    startTimeRef.current = 0;
  };

  const checkLiked = async () => {
    if (!currentTrack) return;
    try {
      const { liked } = await api.checkIfMusicLiked(currentTrack.uuid);
      setIsLiked(liked);
    } catch (e) {
      setIsLiked(false);
    }
  };

  const toggleLike = async () => {
    if (!currentTrack) return;
    try {
      if (isLiked) {
        await api.unlikeMusic(currentTrack.uuid);
        setIsLiked(false);
        toast.success('Removed from liked songs');
      } else {
        await api.likeMusic(currentTrack.uuid);
        setIsLiked(true);
        toast.success('Added to liked songs');
      }
    } catch (error: any) {
      toast.error('Failed to update like status');
    }
  };

  const togglePlay = () => {
    if (!audioRef.current) return;

    if (isPlaying) {
      audioRef.current.pause();
      // Send listen event when manually pausing
      sendListenEvent();
      startTimeRef.current = 0;
    } else {
      audioRef.current.play();
      // Resume tracking
      startTimeRef.current = Date.now();
    }
    setIsPlaying(!isPlaying);
  };

  const handleTimeUpdate = () => {
    if (audioRef.current) {
      setCurrentTime(audioRef.current.currentTime);
    }
  };

  const handleLoadedMetadata = () => {
    if (audioRef.current) {
      setDuration(audioRef.current.duration);
    }
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    const time = parseFloat(e.target.value);
    if (audioRef.current) {
      audioRef.current.currentTime = time;
      setCurrentTime(time);
    }
  };

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const vol = parseFloat(e.target.value);
    setVolume(vol);
    if (audioRef.current) {
      audioRef.current.volume = vol;
    }
  };

  const formatTime = (time: number) => {
    if (isNaN(time)) return '0:00';
    const minutes = Math.floor(time / 60);
    const seconds = Math.floor(time % 60);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  };

  if (!currentTrack) {
    return null;
  }

  return (
    <>
      <audio
        ref={audioRef}
        onTimeUpdate={handleTimeUpdate}
        onLoadedMetadata={handleLoadedMetadata}
        onEnded={() => {
          sendListenEvent();
          playNext();
        }}
      />
      <div className="fixed bottom-0 left-0 right-0 bg-gradient-to-t from-black to-gray-900 border-t border-gray-800 p-4">
        <div className="flex items-center justify-between max-w-screen-2xl mx-auto">
          {/* Track Info */}
          <div className="flex items-center space-x-4 w-1/4">
            {currentTrack.image_path && (
              <img
                src={getFileUrl(currentTrack.image_path)}
                alt={currentTrack.song_name}
                className="w-14 h-14 rounded"
              />
            )}
            <div className="flex-1 min-w-0">
              <p className="text-sm font-semibold truncate">
                {currentTrack.song_name}
              </p>
              <Link href={`/artists/${currentTrack.from_artist}`} className="text-xs text-gray-400 truncate hover:underline block">
                Artist
              </Link>
            </div>
            <button
              onClick={toggleLike}
              className={`${isLiked ? 'text-red-500' : 'text-gray-400'} hover:text-red-400 transition`}
            >
              <Heart className={`w-5 h-5 ${isLiked ? 'fill-current' : ''}`} />
            </button>
          </div>

          {/* Player Controls */}
          <div className="flex flex-col items-center w-2/4">
            <div className="flex items-center space-x-4 mb-2">
              <button
                onClick={playPrevious}
                disabled={currentIndex === 0}
                className="text-gray-400 hover:text-white disabled:opacity-50"
              >
                <SkipBack className="w-5 h-5" />
              </button>
              <button
                onClick={togglePlay}
                className="w-10 h-10 flex items-center justify-center rounded-full bg-white text-black hover:scale-105 transition"
              >
                {isPlaying ? <Pause className="w-5 h-5" /> : <Play className="w-5 h-5 ml-0.5" />}
              </button>
              <button
                onClick={playNext}
                disabled={currentIndex === queue.length - 1}
                className="text-gray-400 hover:text-white disabled:opacity-50"
              >
                <SkipForward className="w-5 h-5" />
              </button>
            </div>
            <div className="flex items-center space-x-2 w-full">
              <span className="text-xs text-gray-400 w-10 text-right">{formatTime(currentTime)}</span>
              <input
                type="range"
                min="0"
                max={duration || 0}
                value={currentTime}
                onChange={handleSeek}
                className="flex-1 h-1 bg-gray-600 rounded-full appearance-none cursor-pointer"
                style={{
                  background: `linear-gradient(to right, #1db954 0%, #1db954 ${(currentTime / duration) * 100}%, #4b5563 ${(currentTime / duration) * 100}%, #4b5563 100%)`,
                }}
              />
              <span className="text-xs text-gray-400 w-10">{formatTime(duration)}</span>
            </div>
          </div>

          {/* Volume Control */}
          <div className="flex items-center justify-end space-x-2 w-1/4">
            <Volume2 className="w-5 h-5 text-gray-400" />
            <input
              type="range"
              min="0"
              max="1"
              step="0.01"
              value={volume}
              onChange={handleVolumeChange}
              className="w-24 h-1 bg-gray-600 rounded-full appearance-none cursor-pointer"
            />
          </div>
        </div>
      </div>
    </>
  );
}
