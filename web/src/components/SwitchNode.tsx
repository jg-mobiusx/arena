import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

function SwitchNodeComponent({ data }: NodeProps) {
  const { label, color } = data as { label: string; color: string; [k: string]: unknown };

  return (
    <>
      <Handle type="target" position={Position.Top} style={{ background: color, width: 6, height: 6, border: '2px solid #ffffff' }} />
      <div className="switch-node" style={{ borderColor: color }}>
        {/* Switch SVG icon — simplified network switch */}
        <svg className="switch-node__icon" viewBox="0 0 32 32" width="24" height="24" fill="none">
          <rect x="2" y="10" width="28" height="12" rx="3" stroke={color} strokeWidth="1.8" />
          {/* Port indicators */}
          <circle cx="8" cy="16" r="2" fill={color} opacity="0.7" />
          <circle cx="14" cy="16" r="2" fill={color} opacity="0.7" />
          <circle cx="20" cy="16" r="2" fill={color} opacity="0.7" />
          <circle cx="26" cy="16" r="2" fill={color} opacity="0.5" />
          {/* Uplink arrows */}
          <path d="M16 10 L16 5" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
          <path d="M13 7 L16 4 L19 7" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
        </svg>
        <span className="switch-node__label" style={{ color }}>{label}</span>
      </div>
      <Handle type="source" position={Position.Right} style={{ background: color, width: 6, height: 6, border: '2px solid #ffffff' }} />
    </>
  );
}

export const SwitchNode = memo(SwitchNodeComponent);
