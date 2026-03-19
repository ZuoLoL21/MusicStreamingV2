'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { formatDate, formatRelativeTime } from '@/lib/formatters';
import { Device } from '@/lib/types';
import { getDeviceId } from '@/lib/deviceId';
import { Monitor, Smartphone, Tablet, Loader2 } from 'lucide-react';
import toast from 'react-hot-toast';
import Cookies from 'js-cookie';
import { useRouter } from 'next/navigation';

export function DeviceList() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [currentDeviceId, setCurrentDeviceId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [revoking, setRevoking] = useState<string | null>(null);
  const [showRevokeAll, setShowRevokeAll] = useState(false);
  const router = useRouter();

  useEffect(() => {
    loadDevices();
    setCurrentDeviceId(getDeviceId());
  }, []);

  const loadDevices = async () => {
    try {
      setLoading(true);
      const deviceList = await api.getDevices();
      setDevices(deviceList);
    } catch (error: any) {
      toast.error('Failed to load devices');
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeDevice = async (device: Device) => {
    const isCurrentDevice = device.device_id === currentDeviceId;
    const message = isCurrentDevice
      ? 'Sign out this device? You will be redirected to login.'
      : `Sign out device "${getDeviceName(device)}"?`;

    if (!window.confirm(message)) return;

    setRevoking(device.uuid);
    try {
      await api.revokeDevice(device.device_id);
      toast.success('Device signed out');

      if (isCurrentDevice) {
        // Clear cookies and redirect to login
        Cookies.remove('token', { path: '/' });
        Cookies.remove('refresh_token', { path: '/' });
        Cookies.remove('user_uuid', { path: '/' });
        router.push('/login');
      } else {
        // Refresh the device list
        await loadDevices();
      }
    } catch (error: any) {
      toast.error('Failed to sign out device');
      console.error(error);
    } finally {
      setRevoking(null);
    }
  };

  const handleRevokeAllDevices = async () => {
    if (!window.confirm('Sign out from all devices? This will log you out.')) {
      setShowRevokeAll(false);
      return;
    }

    setRevoking('all');
    try {
      await api.revokeAllDevices();
      toast.success('All devices signed out');

      // Clear cookies and redirect to login
      Cookies.remove('token', { path: '/' });
      Cookies.remove('refresh_token', { path: '/' });
      Cookies.remove('user_uuid', { path: '/' });
      router.push('/login');
    } catch (error: any) {
      toast.error('Failed to sign out all devices');
      console.error(error);
    } finally {
      setRevoking(null);
      setShowRevokeAll(false);
    }
  };

  const getDeviceName = (device: Device): string => {
    return device.device_name || `Device ${device.device_id.substring(0, 8)}`;
  };

  const getDeviceIcon = (device: Device) => {
    // Simple heuristic - could be enhanced with user agent data
    const name = device.device_name?.toLowerCase() || '';
    if (name.includes('mobile') || name.includes('phone')) {
      return <Smartphone className="w-5 h-5" />;
    } else if (name.includes('tablet') || name.includes('ipad')) {
      return <Tablet className="w-5 h-5" />;
    } else {
      return <Monitor className="w-5 h-5" />;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="w-8 h-8 animate-spin text-gray-400" />
      </div>
    );
  }

  return (
    <div>
      <div className="space-y-3">
        {devices.map((device) => {
          const isCurrentDevice = device.device_id === currentDeviceId;
          const isRevoking = revoking === device.uuid;

          return (
            <div
              key={device.uuid}
              className={`p-4 rounded-lg border ${
                isCurrentDevice
                  ? 'bg-green-900/20 border-green-700'
                  : 'bg-gray-800 border-gray-700'
              }`}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3 flex-1">
                  <div className="text-gray-400 mt-1">{getDeviceIcon(device)}</div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold truncate">{getDeviceName(device)}</h3>
                      {isCurrentDevice && (
                        <span className="px-2 py-0.5 bg-green-600 text-xs rounded-full font-semibold">
                          This device
                        </span>
                      )}
                    </div>
                    <div className="text-sm text-gray-400 mt-1 space-y-0.5">
                      <p>Created: {formatDate(device.created_at)}</p>
                      <p>Last active: {formatRelativeTime(device.last_used_at)}</p>
                      <p>Expires: {formatRelativeTime(device.expires_at)}</p>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleRevokeDevice(device)}
                  disabled={isRevoking || revoking === 'all'}
                  className="ml-4 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded font-semibold transition text-sm"
                >
                  {isRevoking ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Signing out...
                    </span>
                  ) : (
                    'Sign out'
                  )}
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* Sign Out All Devices */}
      {devices.length > 1 && (
        <div className="mt-6 pt-6 border-t border-gray-700">
          {showRevokeAll ? (
            <div className="bg-red-900/20 border border-red-700 rounded-lg p-4">
              <p className="text-sm mb-3">
                This will sign out all devices including this one. You will be redirected to login.
              </p>
              <div className="flex gap-3">
                <button
                  onClick={handleRevokeAllDevices}
                  disabled={revoking !== null}
                  className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 rounded font-semibold transition"
                >
                  {revoking === 'all' ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Signing out...
                    </span>
                  ) : (
                    'Confirm Sign Out All'
                  )}
                </button>
                <button
                  onClick={() => setShowRevokeAll(false)}
                  disabled={revoking !== null}
                  className="px-4 py-2 bg-gray-700 hover:bg-gray-600 disabled:bg-gray-600 rounded font-semibold transition"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <button
              onClick={() => setShowRevokeAll(true)}
              disabled={revoking !== null}
              className="w-full px-4 py-2 bg-gray-800 hover:bg-gray-700 disabled:bg-gray-600 border border-gray-700 rounded font-semibold transition"
            >
              Sign Out All Devices
            </button>
          )}
        </div>
      )}
    </div>
  );
}
