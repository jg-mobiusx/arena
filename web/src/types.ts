export interface DevicePort {
  number: number;
  protocol: string;
  service: string;
}

export interface Device {
  ip: string;
  mac: string;
  hostname: string;
  manufacturer: string;
  os: string;
  open_ports?: DevicePort[];
  first_seen: string;
  last_seen: string;
  is_new: boolean;
  is_online: boolean;
}

export interface VlanMatch {
  manufacturers: string[];
}

export interface VlanConfig {
  id: string;
  name: string;
  tag: number;
  color: string;
  description: string;
  match: VlanMatch;
}

export interface VlansFile {
  vlans: VlanConfig[];
  defaultVlan: VlanConfig;
}
