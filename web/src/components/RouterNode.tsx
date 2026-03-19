import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import { Globe } from 'lucide-react';

function RouterNodeComponent({ data }: NodeProps) {
  const { label, ip } = data as { label: string; ip: string; [k: string]: unknown };

  return (
    <>
      <div className="router-node">
        <div className="router-node__icon">
          <Globe size={22} strokeWidth={1.8} />
        </div>
        <div className="router-node__info">
          <span className="router-node__label">{label}</span>
          <span className="router-node__ip">{ip}</span>
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} style={{ background: '#4f46e5', width: 8, height: 8, border: '2px solid #ffffff' }} />
    </>
  );
}

export const RouterNode = memo(RouterNodeComponent);
