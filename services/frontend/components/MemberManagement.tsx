'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { ArtistMember } from '@/lib/types';
import toast from 'react-hot-toast';

interface MemberManagementProps {
  artistUuid: string;
  currentUserRole: 'owner' | 'manager' | 'member' | null;
}

export default function MemberManagement({
  artistUuid,
  currentUserRole,
}: MemberManagementProps) {
  const [members, setMembers] = useState<ArtistMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddMember, setShowAddMember] = useState(false);
  const [newMemberUsername, setNewMemberUsername] = useState('');
  const [newMemberRole, setNewMemberRole] = useState<'member' | 'manager'>('member');

  useEffect(() => {
    loadMembers();
  }, [artistUuid]);

  const loadMembers = async () => {
    try {
      const data = await api.getArtistMembers(artistUuid);
      setMembers(data);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load members');
    } finally {
      setLoading(false);
    }
  };

  const handleAddMember = async () => {
    if (!newMemberUsername.trim()) {
      toast.error('Please enter a username');
      return;
    }

    try {
      const result = await api.searchUsers(newMemberUsername.trim());
      const user = result.users.find((u) => u.username.toLowerCase() === newMemberUsername.trim().toLowerCase());

      if (!user) {
        toast.error('User not found');
        return;
      }

      await api.addMemberToArtist(artistUuid, user.uuid, newMemberRole);
      toast.success('Member added successfully');
      loadMembers();
      setShowAddMember(false);
      setNewMemberUsername('');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to add member');
    }
  };

  const handleRemoveMember = async (userUuid: string, username: string) => {
    if (!confirm(`Remove ${username} from this artist?`)) {
      return;
    }

    try {
      await api.removeMemberFromArtist(artistUuid, userUuid);
      toast.success('Member removed successfully');
      loadMembers();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to remove member');
    }
  };

  const handleChangeRole = async (userUuid: string, username: string, currentRole: string) => {
    const newRole = prompt(
      `Change role for ${username}.\nCurrent: ${currentRole}\nEnter new role (owner/manager/member):`
    );

    if (!newRole || !['owner', 'manager', 'member'].includes(newRole.toLowerCase())) {
      return;
    }

    try {
      await api.changeArtistMemberRole(artistUuid, userUuid, newRole.toLowerCase());
      toast.success('Role updated successfully');
      loadMembers();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update role');
    }
  };

  const canManageMembers = currentUserRole === 'owner';
  const canChangeRoles = currentUserRole === 'owner';

  if (loading) {
    return <div className="text-center py-8">Loading members...</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Members</h2>
        {canManageMembers && (
          <button
            onClick={() => setShowAddMember(!showAddMember)}
            className="bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold"
          >
            {showAddMember ? 'Cancel' : 'Add Member'}
          </button>
        )}
      </div>

      {/* Add Member Form */}
      {showAddMember && canManageMembers && (
        <div className="bg-gray-700 rounded-lg p-4 space-y-4">
          <div>
            <label className="block text-sm font-semibold mb-2">Username</label>
            <input
              type="text"
              value={newMemberUsername}
              onChange={(e) => setNewMemberUsername(e.target.value)}
              className="w-full px-4 py-2 bg-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter username"
            />
          </div>
          <div>
            <label className="block text-sm font-semibold mb-2">Role</label>
            <select
              value={newMemberRole}
              onChange={(e) => setNewMemberRole(e.target.value as 'member' | 'manager')}
              className="w-full px-4 py-2 bg-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="member">Member</option>
              <option value="manager">Manager</option>
            </select>
          </div>
          <button
            onClick={handleAddMember}
            className="w-full bg-blue-600 hover:bg-blue-700 px-4 py-2 rounded-lg font-semibold"
          >
            Add Member
          </button>
        </div>
      )}

      {/* Members List */}
      <div className="space-y-3">
        {members.map((member) => (
          <div
            key={member.user_uuid}
            className="flex items-center justify-between p-4 bg-gray-800 rounded-lg"
          >
            <div className="flex-1">
              <h3 className="font-semibold">{member.username}</h3>
              <p className="text-sm text-gray-400 capitalize">{member.role}</p>
              <p className="text-xs text-gray-500">
                Added {new Date(member.added_at).toLocaleDateString()}
              </p>
            </div>

            {canManageMembers && member.role !== 'owner' && (
              <div className="flex gap-2">
                {canChangeRoles && (
                  <button
                    onClick={() =>
                      handleChangeRole(member.user_uuid, member.username, member.role)
                    }
                    className="bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded text-sm font-semibold"
                  >
                    Change Role
                  </button>
                )}
                <button
                  onClick={() => handleRemoveMember(member.user_uuid, member.username)}
                  className="bg-red-600 hover:bg-red-700 px-3 py-1 rounded text-sm font-semibold"
                >
                  Remove
                </button>
              </div>
            )}
          </div>
        ))}

        {members.length === 0 && (
          <div className="text-center py-8 text-gray-400">No members yet</div>
        )}
      </div>
    </div>
  );
}
