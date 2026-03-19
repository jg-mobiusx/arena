import { memo } from 'react';
import type { NodeProps } from '@xyflow/react';
import type { LucideIcon } from 'lucide-react';

interface GroupNodeData {
  label: string;
  count: number;
  icon: LucideIcon;
  color: string;
  [key: string]: unknown;
}

function GroupNodeComponent({ data }: NodeProps) {
  const { label, count, icon: Icon, color } = data as unknown as GroupNodeData;

  return (
    <div className="group-node">
      <div className="group-node__header">
        <div className="group-node__icon" style={{ color }}>
          <Icon size={16} />
        </div>
        <span className="group-node__label">{label}</span>
        <span className="group-node__count">{count}</span>
      </div>
    </div>
  );
}

export const GroupNode = memo(GroupNodeComponent);
