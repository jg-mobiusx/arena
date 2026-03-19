import type { Device, VlanConfig, VlansFile } from '../types';
import type { Node, Edge } from '@xyflow/react';
import {
  Wifi,
  Monitor,
  Cpu,
  Camera,
  Speaker,
  Tv,
  Printer,
  Router,
  Lightbulb,
  Smartphone,
  Server,
  HardDrive,
  Globe,
  type LucideIcon,
} from 'lucide-react';

// ─── Device Icon Mapping ─────────────────────────────────────────────
// Maps manufacturers (and hostnames where useful) to appropriate network icons.

const MANUFACTURER_ICON: Record<string, LucideIcon> = {
  'Apple': Monitor,
  'Google': Tv,
  'Sonos': Speaker,
  'Raspberry Pi Trading': Cpu,
  'Raspberry Pi Foundation': Cpu,
  'Hangzhou Hikvision Digital Technology': Camera,
  'Espressif': Lightbulb,
  'Netgear': Wifi,
  'Arcadyan': Router,
  'Belkin International': Wifi,
  'Hewlett Packard': Printer,
  'LG Innotek': Tv,
  'LightwaveRF Technology': Lightbulb,
  'Microchip Technology': Lightbulb,
  'AzureWave Technology': Wifi,
};

const HOSTNAME_ICON: Record<string, LucideIcon> = {
  'macmini': Server,
  'macbookpro': Monitor,
  'octopi': Server,
  'pi.hole': Server,
  'emonpi': Cpu,
  'chromecast': Tv,
  'neo-hub': Lightbulb,
  'office-apple-tv': Tv,
  'lgwebostv': Tv,
};

export function getDeviceIcon(device: Device): LucideIcon {
  // Check hostname first (more specific)
  const hn = device.hostname.toLowerCase();
  for (const [key, icon] of Object.entries(HOSTNAME_ICON)) {
    if (hn.includes(key)) return icon;
  }
  // Then manufacturer
  return MANUFACTURER_ICON[device.manufacturer] || HardDrive;
}

// ─── VLAN Assigner ───────────────────────────────────────────────────
function assignVlan(device: Device, vlansFile: VlansFile): VlanConfig {
  for (const vlan of vlansFile.vlans) {
    if (vlan.match.manufacturers.includes(device.manufacturer)) {
      return vlan;
    }
  }
  return vlansFile.defaultVlan;
}

// ─── Layout Constants ────────────────────────────────────────────────
const MAX_ROW_WIDTH = 1200;           // wrap devices after this width
const BACKBONE_Y = 80;                // Y position of the backbone line within a VLAN band
const DEVICE_DROP_Y = 60;             // vertical drop distance from backbone to device
const DEVICE_WIDTH = 154;
const DEVICE_SPACING_X = 166;         // horizontal spacing between devices
const DEVICE_ROW_SPACING_Y = 110;     // vertical spacing between rows of devices
const VLAN_PADDING_X = 40;
const VLAN_GAP_Y = 50;               // gap between VLAN bands
const CANVAS_OFFSET_X = 40;
const CANVAS_OFFSET_Y = 40;
const SWITCH_NODE_WIDTH = 56;
const BACKBONE_EXTRA = 80;           // extra backbone length past last column

// ─── Build the Bus Topology Graph ────────────────────────────────────
export function buildGraph(
  devices: Device[],
  vlansFile: VlansFile
): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = [];
  const edges: Edge[] = [];

  // 1. Group devices by VLAN
  const vlanGroups = new Map<string, { vlan: VlanConfig; devices: Device[] }>();

  // Pre-create groups in VLAN order so they render top-to-bottom
  for (const vlan of vlansFile.vlans) {
    vlanGroups.set(vlan.id, { vlan, devices: [] });
  }
  vlanGroups.set(vlansFile.defaultVlan.id, { vlan: vlansFile.defaultVlan, devices: [] });

  for (const device of devices) {
    const vlan = assignVlan(device, vlansFile);
    vlanGroups.get(vlan.id)!.devices.push(device);
  }

  // 2. Build each VLAN as a horizontal bus
  let currentY = CANVAS_OFFSET_Y;

  for (const [vlanId, group] of vlanGroups) {
    if (group.devices.length === 0) continue;

    const vlan = group.vlan;
    const deviceCount = group.devices.length;

    // Calculate grid layout for devices
    const devicesPerRow = Math.max(1, Math.floor((MAX_ROW_WIDTH - VLAN_PADDING_X - SWITCH_NODE_WIDTH) / DEVICE_SPACING_X));
    const maxColumns = Math.min(deviceCount, devicesPerRow);
    const numRows = Math.ceil(deviceCount / devicesPerRow);

    // Calculate VLAN band dimensions
    const backboneLength = VLAN_PADDING_X + (maxColumns * DEVICE_SPACING_X) + BACKBONE_EXTRA;
    const bandHeight = BACKBONE_Y + DEVICE_DROP_Y + (numRows * DEVICE_ROW_SPACING_Y);

    // ── VLAN Zone (parent group node) ──
    const vlanNodeId = `vlan-${vlanId}`;
    nodes.push({
      id: vlanNodeId,
      type: 'vlanZone',
      position: { x: CANVAS_OFFSET_X, y: currentY },
      data: {
        label: `VLAN ${vlan.tag}: ${vlan.name}`,
        description: vlan.description,
        color: vlan.color,
        count: deviceCount,
      },
      style: {
        width: backboneLength + VLAN_PADDING_X,
        height: bandHeight,
      },
      selectable: false,
      draggable: false,
    });

    // ── Switch / Router node at the left of the backbone ──
    const switchId = `switch-${vlanId}`;
    const switchX = VLAN_PADDING_X;
    const switchY = BACKBONE_Y - (SWITCH_NODE_WIDTH / 2);

    nodes.push({
      id: switchId,
      type: 'switchNode',
      position: { x: switchX, y: switchY },
      parentId: vlanNodeId,
      extent: 'parent' as const,
      data: {
        label: vlan.tag === 0 ? 'Unmanaged' : `SW-${vlan.tag}`,
        color: vlan.color,
      },
    });

    // ── Backbone anchor at the right end ──
    const anchorId = `anchor-${vlanId}`;
    const anchorX = VLAN_PADDING_X + (maxColumns * DEVICE_SPACING_X) + BACKBONE_EXTRA - 20;

    nodes.push({
      id: anchorId,
      type: 'backboneAnchor',
      position: { x: anchorX, y: BACKBONE_Y - 2 },
      parentId: vlanNodeId,
      extent: 'parent' as const,
      data: { color: vlan.color },
    });

    // ── Backbone edge (switch → anchor) ──
    edges.push({
      id: `backbone-${vlanId}`,
      source: switchId,
      target: anchorId,
      type: 'straight',
      style: {
        stroke: vlan.color,
        strokeWidth: 3,
        strokeOpacity: 0.6,
      },
      selectable: false,
      data: { isBackbone: true },
    });

    // ── Device nodes mapped onto a grid below the backbone ──
    const sorted = [...group.devices].sort((a, b) => {
      if (a.is_online !== b.is_online) return a.is_online ? -1 : 1;
      return a.ip.localeCompare(b.ip, undefined, { numeric: true });
    });

    // Create a single backbone tap point for each vertical column
    for (let col = 0; col < maxColumns; col++) {
      const tapId = `tap-${vlanId}-col-${col}`;
      const tapX = VLAN_PADDING_X + SWITCH_NODE_WIDTH + 30 + (col * DEVICE_SPACING_X);
      
      nodes.push({
        id: tapId,
        type: 'backboneAnchor',
        position: {
          x: tapX + (DEVICE_WIDTH / 2) - 2,
          y: BACKBONE_Y - 2,
        },
        parentId: vlanNodeId,
        extent: 'parent' as const,
        data: { color: vlan.color },
      });
    }

    // Place devices and drop them vertically from their column's tap point
    sorted.forEach((device, i) => {
      const deviceId = `device-${device.ip.replace(/\./g, '-')}`;
      const col = i % devicesPerRow;
      const row = Math.floor(i / devicesPerRow);
      
      const deviceX = VLAN_PADDING_X + SWITCH_NODE_WIDTH + 30 + (col * DEVICE_SPACING_X);
      const deviceY = BACKBONE_Y + DEVICE_DROP_Y + (row * DEVICE_ROW_SPACING_Y);
      const tapId = `tap-${vlanId}-col-${col}`;

      // ── Device node ──
      nodes.push({
        id: deviceId,
        type: 'deviceNode',
        position: { x: deviceX, y: deviceY },
        parentId: vlanNodeId,
        extent: 'parent' as const,
        data: {
          device,
          color: vlan.color,
          icon: getDeviceIcon(device),
        },
      });

      // ── Vertical connection edges ──
      // First row connects to the backbone tap. Subsequent rows connect to the device above them (daisy chain visual).
      const sourceId = row === 0 ? tapId : `device-${sorted[i - devicesPerRow].ip.replace(/\./g, '-')}`;

      edges.push({
        id: `link-${deviceId}`,
        source: sourceId,
        target: deviceId,
        type: 'straight',
        style: {
          stroke: vlan.color,
          strokeWidth: 1.5,
          strokeOpacity: device.is_online ? 0.5 : 0.2,
          strokeDasharray: device.is_online ? undefined : '4 4',
        },
        animated: device.is_online,
      });
    });

    currentY += bandHeight + VLAN_GAP_Y;
  }

  // 3. Add a top-level router node connected to all VLAN switches
  const routerId = 'router-gateway';
  nodes.push({
    id: routerId,
    type: 'routerNode',
    position: { x: CANVAS_OFFSET_X + VLAN_PADDING_X, y: CANVAS_OFFSET_Y - 80 },
    data: {
      label: 'Gateway',
      ip: '172.16.44.1',
    },
  });

  // Connect gateway to each VLAN switch
  for (const [vlanId, group] of vlanGroups) {
    if (group.devices.length === 0) continue;
    const switchId = `switch-${vlanId}`;
    edges.push({
      id: `uplink-${vlanId}`,
      source: routerId,
      target: switchId,
      type: 'smoothstep',
      style: {
        stroke: group.vlan.color,
        strokeWidth: 2,
        strokeOpacity: 0.35,
      },
    });
  }

  return { nodes, edges };
}

// Re-export for use in components
export { Globe, Router, Smartphone, Server, HardDrive, Wifi };
