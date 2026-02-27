'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import { User } from '@/lib/types';
import toast from 'react-hot-toast';

interface ProfileEditProps {
  user: User;
  onClose: () => void;
  onUpdate: () => void;
}

export default function ProfileEdit({ user, onClose, onUpdate }: ProfileEditProps) {
  const [tab, setTab] = useState<'profile' | 'email' | 'password' | 'image'>('profile');
  const [loading, setLoading] = useState(false);

  // Profile fields
  const [username, setUsername] = useState(user.username);
  const [bio, setBio] = useState(user.bio || '');

  // Email fields
  const [newEmail, setNewEmail] = useState('');
  const [emailPassword, setEmailPassword] = useState('');

  // Password fields
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  // Image field
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);

  const handleProfileUpdate = async () => {
    setLoading(true);
    try {
      await api.updateProfile(username, bio || undefined);
      toast.success('Profile updated successfully');
      onUpdate();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  const handleEmailUpdate = async () => {
    if (!newEmail || !emailPassword) {
      toast.error('Please fill in all fields');
      return;
    }
    setLoading(true);
    try {
      await api.updateEmail(emailPassword, newEmail);
      toast.success('Email updated successfully');
      setNewEmail('');
      setEmailPassword('');
      onUpdate();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update email');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordUpdate = async () => {
    if (!oldPassword || !newPassword || !confirmPassword) {
      toast.error('Please fill in all fields');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('Passwords do not match');
      return;
    }
    setLoading(true);
    try {
      await api.updatePassword(oldPassword, newPassword);
      toast.success('Password updated successfully');
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update password');
    } finally {
      setLoading(false);
    }
  };

  const handleImageUpdate = async () => {
    if (!imageFile) {
      toast.error('Please select an image');
      return;
    }
    if (imageFile.size > 10 * 1024 * 1024) {
      toast.error('Image must be less than 10MB');
      return;
    }
    setLoading(true);
    try {
      await api.uploadProfileImage(imageFile);
      toast.success('Profile image updated successfully');
      setImageFile(null);
      setImagePreview(null);
      onUpdate();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to upload image');
    } finally {
      setLoading(false);
    }
  };

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setImageFile(file);
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-800 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-bold">Edit Profile</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-white text-2xl"
            >
              &times;
            </button>
          </div>

          {/* Tabs */}
          <div className="flex gap-2 mb-6 border-b border-gray-700">
            {(['profile', 'email', 'password', 'image'] as const).map((t) => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`px-4 py-2 font-semibold capitalize ${
                  tab === t
                    ? 'text-blue-500 border-b-2 border-blue-500'
                    : 'text-gray-400 hover:text-white'
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          {/* Profile Tab */}
          {tab === 'profile' && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-semibold mb-2">Username</label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">Bio</label>
                <textarea
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  rows={4}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Tell us about yourself..."
                />
              </div>
              <button
                onClick={handleProfileUpdate}
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
              >
                {loading ? 'Updating...' : 'Update Profile'}
              </button>
            </div>
          )}

          {/* Email Tab */}
          {tab === 'email' && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-semibold mb-2">Current Email</label>
                <input
                  type="text"
                  value={user.email}
                  disabled
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg opacity-50"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">New Email</label>
                <input
                  type="email"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">Confirm Password</label>
                <input
                  type="password"
                  value={emailPassword}
                  onChange={(e) => setEmailPassword(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <button
                onClick={handleEmailUpdate}
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
              >
                {loading ? 'Updating...' : 'Update Email'}
              </button>
            </div>
          )}

          {/* Password Tab */}
          {tab === 'password' && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-semibold mb-2">Old Password</label>
                <input
                  type="password"
                  value={oldPassword}
                  onChange={(e) => setOldPassword(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">New Password</label>
                <input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">Confirm New Password</label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <button
                onClick={handlePasswordUpdate}
                disabled={loading}
                className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
              >
                {loading ? 'Updating...' : 'Update Password'}
              </button>
            </div>
          )}

          {/* Image Tab */}
          {tab === 'image' && (
            <div className="space-y-4">
              <div className="text-center">
                <img
                  src={imagePreview || user.profile_image_path || '/default-avatar.png'}
                  alt="Profile preview"
                  className="w-32 h-32 rounded-full object-cover mx-auto mb-4"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-2">Select Image</label>
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleImageSelect}
                  className="w-full px-4 py-2 bg-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-400 mt-1">Max 10MB</p>
              </div>
              <button
                onClick={handleImageUpdate}
                disabled={loading || !imageFile}
                className="w-full bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-semibold disabled:opacity-50"
              >
                {loading ? 'Uploading...' : 'Upload Image'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
