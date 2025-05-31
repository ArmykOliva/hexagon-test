// --- TypeScript Interfaces ---

export interface Subnet {
	name: string;
	country: string;
	address: string;
	scope: string;
	type: string; // 'ipv4' | 'ipv6'
	prefix: number;
}

export interface NetworkInterface {
	name: string;
	subnets: Subnet[];
	mac?: string; 
	ip4?: string[]; 
	ip6?: string[]; 
}

export interface Distribution {
	name: string;
	version: string;
	kernel: string;
	kernel_architecture?: string;
	kernel_modules?: string[];
	kernel_version?: string;
	manager?: string;
	vendor?: string;
}

export interface Fingerprint {
	country: string;
	locale: string;
	timezone: string;
	token?: string;
}

export interface BiosInfo {
	name: string;
	bus: string;
	system: any; 
	subsystem: any; 
}

export interface BoardInfo {
	name: string;
	bus: string;
	system: any;
	subsystem: any;
}

export interface DeviceSystemInfo {
	device: string;
	vendor: string;
	name: string;
}

export interface Device {
	name: string;
	bus: string;
	system: DeviceSystemInfo;
	subsystem?: DeviceSystemInfo | null;
}

export interface ServiceUser {
	name: string;
	password?: string; 
	type: string;
}

export interface ServiceManager {
	name: string;
}

export interface SocketInfo {
	host: string;
	port: number;
}

export interface ServiceConnection {
	socket: SocketInfo;
	protocol: string; 
	type: string; 
}

export interface Service {
	pid: number;
	name: string;
	command: string;
	arguments: string[];
	folder: string;
	environment: Record<string, string>;
	user: ServiceUser; 
	manager: ServiceManager; 
	filesystem: string[];
	connections: ServiceConnection[]; 
	dependencies: any[]; 
	packages: any[]; 
}

export interface PackageInfo {
	name: string;
	version: string;
	architecture: string;
	manager: string;
	vendor?: string; 
	url: string;
	datetime?: string; 
	maintainers?: string[]; 
	filesystem?: string[]; 
	conflicts?: string[]; 
	dependencies?: any[]; 
	provides?: string[]; 
	replaces?: string[]; 
	unresolved?: string[]; 
}

export interface UserGroup {
	id: number;
	name: string;
	password?: string;
	type?: string;
}

export interface User {
	id: number;
	name: string;
	password?: string;
	folder?: string;
	groups?: UserGroup[];
	shell?: string;
	type?: string; 
}

// --- Interfaces for System-to-System Connections ---
export interface SystemConnectionSocket {
	host: string; // System ID or resolvable hostname/IP
	port: number;
}

export interface SystemToSystemConnection {
	source: SystemConnectionSocket;
	target: SystemConnectionSocket;
	protocol: string; // e.g., "tcp", "udp"
	type: string; // e.g., "established", "related"
}

export interface System {
	id: string; 
	name: string; 
	hostname: string;
	datetime: string;
	distribution: Distribution;
	fingerprint: Fingerprint;
	bios: BiosInfo | null;
	board: BoardInfo | null;
	devices: Device[] | null;
	drives: any | null; 
	incidents: any | null; 
	mitigations: any[] | null;
	networks: NetworkInterface[]; 
	packages: PackageInfo[] | null; 
	programs: any | null; 
	responses?: any[]; 
	services: Service[] | null;
	antiques?: any[]; 
	updates: PackageInfo[] | null; 
	users?: User[] | null;
	connections?: SystemToSystemConnection[] | null; // Added for system-to-system links
} 