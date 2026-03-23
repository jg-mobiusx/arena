import { useEffect, useState } from 'react';
import { NetworkGraph } from './components/NetworkGraph';
import type { Device, VlansFile, SystemStatus } from './types';

export default function App() {
  const [devices, setDevices] = useState<Device[] | null>(null);
  const [vlansConfig, setVlansConfig] = useState<VlansFile | null>(null);
  const [systemStatus, setSystemStatus] = useState<SystemStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    async function fetchData() {
      try {
        const [devicesRes, vlansRes, statusRes] = await Promise.all([
          fetch('/api/devices.json'),
          fetch('/api/vlans.json'),
          fetch('/api/status')
        ]);

        if (!devicesRes.ok) throw new Error(`Failed to fetch devices: ${devicesRes.status}`);
        if (!vlansRes.ok) throw new Error(`Failed to fetch VLANs: ${vlansRes.status}`);
        if (!statusRes.ok) throw new Error(`Failed to fetch status: ${statusRes.status}`);

        const devicesData = await devicesRes.json();
        const vlansData = await vlansRes.json();
        const statusData = await statusRes.json();

        if (mounted) {
          setDevices(devicesData);
          setVlansConfig(vlansData);
          setSystemStatus(statusData);
          setError(null);
        }
      } catch (err: unknown) {
        console.error('Data fetch error:', err);
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Unknown error fetching data');
        }
      }
    }

    // Initial fetch
    fetchData();

    // Poll every 10 seconds for live updates
    const interval = setInterval(fetchData, 10000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  if (error) {
    return (
      <div style={{ padding: 20, color: 'var(--accent-rose)', fontFamily: 'var(--font-sans)', textAlign: 'center' }}>
        <h2>Connection Error</h2>
        <p>{error}</p>
        <p style={{ fontSize: 14, color: 'var(--text-muted)' }}>Is the Arena daemon running?</p>
      </div>
    );
  }

  if (!devices || !vlansConfig || !systemStatus) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', fontFamily: 'var(--font-sans)', color: 'var(--text-secondary)' }}>
        <div>Loading network topology...</div>
      </div>
    );
  }

  return <NetworkGraph devices={devices} vlansConfig={vlansConfig} systemStatus={systemStatus} />;
}
