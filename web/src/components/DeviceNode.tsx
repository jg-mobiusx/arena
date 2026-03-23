import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import type { Device } from '../types';
import type { LucideIcon } from 'lucide-react';

interface DeviceNodeData {
  device: Device;
  color: string;
  icon: LucideIcon;
  [key: string]: unknown;
}

function DeviceNodeComponent({ data }: NodeProps) {
  const { device, color, icon: Icon } = data as unknown as DeviceNodeData;
  const statusClass = device.is_new ? 'new' : device.is_online ? 'online' : 'offline';
  const displayName = device.hostname || device.manufacturer || device.ip;

  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: color, width: 5, height: 5, border: '2px solid #ffffff', opacity: 0 }} />
      <div className={`device-node device-node--${statusClass}`}>
        <div className="device-node__row">
          <div className="device-node__icon" style={{ color }}>
            <Icon size={16} strokeWidth={1.8} />
          </div>
          <span className="device-node__hostname" title={displayName}>
            {displayName}
          </span>
          <span className={`device-node__status-dot device-node__status-dot--${device.is_online ? 'online' : 'offline'}`} />
        </div>
        <div className="device-node__ip">{device.ip}</div>
        <div className="device-node__meta">
          <span className="device-node__mac">{device.mac || '—'}</span>
          {device.is_new && <span className="device-node__badge device-node__badge--new">NEW</span>}
        </div>
        
        {(device.http_title || device.server_header) && (
          <div className="device-node__http-info" style={{ marginTop: 6, padding: '4px 6px', background: 'var(--surface-hover)', borderRadius: 4, fontSize: 10, color: 'var(--text-secondary)' }}>
            {device.http_title && <div style={{ fontWeight: 500, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{device.http_title}</div>}
            {device.server_header && <div style={{ fontSize: 9, opacity: 0.8 }}>{device.server_header}</div>}
          </div>
        )}

        {device.open_ports && device.open_ports.length > 0 && (
          <div className="device-node__ports" style={{ marginTop: 6 }}>
            {device.open_ports.map((p) => (
              <span key={p.number} className="device-node__port-tag">
                :{p.number}
              </span>
            ))}
          </div>
        )}
      </div>
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </>
  );
}

export const DeviceNode = memo(DeviceNodeComponent);
