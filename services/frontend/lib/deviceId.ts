import Cookies from 'js-cookie';

const DEVICE_ID_COOKIE = 'device_id';

/**
 * Generate a unique device ID using crypto.randomUUID() or fallback
 */
function generateDeviceId(): string {
  // Use crypto.randomUUID if available (modern browsers)
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID();
  }

  // Fallback: generate a UUID v4-like string
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

/**
 * Get or create a device ID stored in cookies
 */
export function getDeviceId(): string {
  // Check if device ID already exists in cookies
  let deviceId = Cookies.get(DEVICE_ID_COOKIE);

  if (!deviceId) {
    // Generate new device ID
    deviceId = generateDeviceId();

    // Store in cookie (expires in 10 years - essentially permanent)
    Cookies.set(DEVICE_ID_COOKIE, deviceId, {
      expires: 3650, // 10 years
      path: '/',
      sameSite: 'lax',
      secure: false, // Allow on localhost
    });
  }

  return deviceId;
}

/**
 * Clear the device ID (useful for logout/reset)
 */
export function clearDeviceId(): void {
  Cookies.remove(DEVICE_ID_COOKIE, { path: '/' });
}
