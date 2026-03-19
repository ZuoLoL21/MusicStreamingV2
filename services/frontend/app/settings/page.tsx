'use client';

import { DeviceList } from '@/components/DeviceList';
import { Settings as SettingsIcon } from 'lucide-react';

export default function SettingsPage() {
  return (
    <div className="p-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-3 mb-8">
          <SettingsIcon className="w-8 h-8" />
          <h1 className="text-4xl font-bold">Settings</h1>
        </div>

        {/* Device Management Section */}
        <section className="mb-12">
          <h2 className="text-2xl font-bold mb-2">Manage Devices</h2>
          <p className="text-gray-400 mb-6">
            View and manage devices where you're currently signed in. You can sign out from devices you no longer use.
          </p>
          <DeviceList />
        </section>

        {/* Future sections can be added here */}
        {/* Example: Notifications, Privacy, Account, etc. */}
      </div>
    </div>
  );
}
