import { useMemo, useCallback, useState, useEffect } from 'react';
import {
  ReactFlow,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  useNodesState,
  useEdgesState,
  type ColorMode,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import type { Device, VlansFile, SystemStatus } from '../types';
import { DeviceNode } from './DeviceNode';
import { GroupNode } from './GroupNode';
import { RouterNode } from './RouterNode';
import { SwitchNode } from './SwitchNode';
import { BackboneAnchor } from './BackboneAnchor';
import { VlanZone } from './VlanZone';
import { buildGraph } from './graphBuilder';

const nodeTypes = {
  deviceNode: DeviceNode,
  groupNode: GroupNode,
  routerNode: RouterNode,
  switchNode: SwitchNode,
  backboneAnchor: BackboneAnchor,
  vlanZone: VlanZone,
};

interface NetworkGraphProps {
  devices: Device[];
  vlansConfig: VlansFile;
  systemStatus: SystemStatus;
}

export function NetworkGraph({ devices, vlansConfig, systemStatus }: NetworkGraphProps) {
  const [filter, setFilter] = useState<'all' | 'online' | 'offline'>('all');

  const filtered = useMemo(() => {
    if (filter === 'online') return devices.filter((d) => d.is_online);
    if (filter === 'offline') return devices.filter((d) => !d.is_online);
    return devices;
  }, [devices, filter]);

  const { nodes: initialNodes, edges: initialEdges } = useMemo(
    () => buildGraph(filtered, vlansConfig),
    [filtered, vlansConfig]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  useEffect(() => {
    setNodes(initialNodes);
    setEdges(initialEdges);
  }, [initialNodes, initialEdges, setNodes, setEdges]);

  const onlineCount = devices.filter((d) => d.is_online).length;
  const offlineCount = devices.filter((d) => !d.is_online).length;

  const minimapNodeColor = useCallback(
    (node: { type?: string }) => {
      if (node.type === 'vlanZone') return '#e2e8f0';
      if (node.type === 'routerNode') return '#4f46e5';
      if (node.type === 'switchNode') return '#059669';
      if (node.type === 'backboneAnchor') return 'transparent';
      return '#cbd5e1';
    },
    []
  );

  return (
    <div className="app-container">
      {/* ── Header ── */}
      <header className="app-header">
        <div className="app-header__brand">
          <div className="app-header__logo">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10" />
              <path d="M2 12h20" />
            </svg>
          </div>
          <div>
            <div className="app-header__title">Arena</div>
            <div className="app-header__subtitle">Network Topology</div>
          </div>
        </div>

        <div className="app-header__stats">
          <div className="stat-badge">
            <span className="stat-badge__dot" style={{ background: systemStatus.ping_ms > 0 ? '#10b981' : '#ef4444' }} />
            WAN {systemStatus.ping_ms > 0 ? `${systemStatus.ping_ms}ms` : 'Offline'}
          </div>
          <div className="stat-badge" style={{ color: 'var(--text-muted)', background: 'transparent', border: 'none', paddingLeft: 0 }}>
            Scanned {new Date(systemStatus.last_scan).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </div>
          <div className="stat-badge" style={{ marginLeft: 'auto' }}>
            <span className="stat-badge__dot stat-badge__dot--total" />
            {devices.length} devices
          </div>
          <div className="stat-badge">
            <span className="stat-badge__dot stat-badge__dot--online" />
            {onlineCount} online
          </div>
          <div className="stat-badge">
            <span className="stat-badge__dot stat-badge__dot--offline" />
            {offlineCount} offline
          </div>
          <div className="stat-badge">
            <span className="stat-badge__dot" style={{ background: '#a78bfa' }} />
            {vlansConfig.vlans.length} VLANs
          </div>
        </div>

        <div className="app-header__controls">
          {(['all', 'online', 'offline'] as const).map((f) => (
            <button
              key={f}
              className={`control-btn ${filter === f ? 'control-btn--active' : ''}`}
              onClick={() => setFilter(f)}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
        </div>
      </header>

      {/* ── Canvas ── */}
      <div className="flow-canvas">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          nodeTypes={nodeTypes}
          fitView
          fitViewOptions={{ padding: 0.12 }}
          minZoom={0.1}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable={false}
          panOnDrag={true}
          selectionOnDrag={false}
          zoomOnScroll={true}
          zoomOnPinch={true}
          zoomOnDoubleClick={true}
          panOnScroll={true}
          preventScrolling={false}
          proOptions={{ hideAttribution: true }}
          colorMode={'light' as ColorMode}
        >
          <Background variant={BackgroundVariant.Dots} gap={24} size={1} color="#e2e8f0" />
          <Controls showInteractive={false} />
          <MiniMap
            nodeColor={minimapNodeColor}
            maskColor="rgba(248, 250, 252, 0.8)"
            pannable
            zoomable
          />
        </ReactFlow>
      </div>
    </div>
  );
}
