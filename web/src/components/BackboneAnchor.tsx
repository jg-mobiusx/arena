import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

function BackboneAnchorComponent({ data }: NodeProps) {
  const { color } = data as { color: string; [k: string]: unknown };

  return (
    <>
      <Handle type="target" position={Position.Left} style={{ opacity: 0 }} />
      <div className="backbone-anchor" style={{ background: color }} />
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </>
  );
}

export const BackboneAnchor = memo(BackboneAnchorComponent);
