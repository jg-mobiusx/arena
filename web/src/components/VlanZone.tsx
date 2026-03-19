import { memo } from 'react';
import type { NodeProps } from '@xyflow/react';

function VlanZoneComponent({ data }: NodeProps) {
  const { label, description, color, count } = data as {
    label: string;
    description: string;
    color: string;
    count: number;
    [k: string]: unknown;
  };

  return (
    <div className="vlan-zone" style={{ borderColor: `${color}33` }}>
      <div className="vlan-zone__header">
        <div className="vlan-zone__indicator" style={{ background: color }} />
        <span className="vlan-zone__label" style={{ color }}>{label}</span>
        <span className="vlan-zone__description">{description}</span>
        <span className="vlan-zone__count">{count} devices</span>
      </div>
    </div>
  );
}

export const VlanZone = memo(VlanZoneComponent);
