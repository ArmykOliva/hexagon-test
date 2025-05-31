<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Panzoom from 'panzoom'; // Import Panzoom
	import type { PanzoomObject } from 'panzoom'; // Import PanzoomObject type
	import Hexagon from '../lib/components/Hexagon.svelte'; // Import Hexagon component

	// Define interfaces based on the provided JSON structure (assuming these are still valid)
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
		// Older fields like mac, ip4, ip6 might be part of a more detailed view or sub-object if available
		// For now, focusing on the provided 'subnets' structure
		mac?: string; // Optional, if still relevant from older data or if it can be derived
		ip4?: string[]; // Optional
		ip6?: string[]; // Optional
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
		system: any; // Can be more specific if structure is known
		subsystem: any; // Can be more specific
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
	
	// Assuming 'services' if not null, would be an array of objects with at least a name
	// This is a placeholder; update when actual service object structure is known
	export interface Service {
		pid: number;
		name: string;
		command: string;
		arguments: string[];
		folder: string;
		environment: Record<string, string>;
		user: ServiceUser; // Define ServiceUser
		manager: ServiceManager; // Define ServiceManager
		filesystem: string[];
		connections: ServiceConnection[]; // Define ServiceConnection
		dependencies: any[]; // Can be more specific if structure known
		packages: any[]; // Can be more specific if structure known
	}

	export interface ServiceUser {
		name: string;
		password?: string; // Or just type: string if always present
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
		protocol: string; // e.g., "tcp"
		type: string; // e.g., "server"
	}

	// For the 'packages' field (if it's ever populated with old structure)
	// and for the 'updates' field
	export interface PackageInfo { // Renamed from Package to avoid conflict if structure differs
		name: string;
		version: string;
		architecture: string;
		manager: string;
		vendor?: string; // Was in old 'Package', might be in 'updates' items
		url: string;
		datetime?: string; // Was in old 'Package'
		maintainers?: string[]; // Was in old 'Package'
		filesystem?: string[]; // Was in old 'Package'
		conflicts?: string[]; // Was in old 'Package'
		dependencies?: any[]; // Was in old 'Package', use 'any' or define PackageDependency
		provides?: string[]; // Was in old 'Package'
		replaces?: string[]; // Was in old 'Package'
		unresolved?: string[]; // Was in old 'Package'
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
		type?: string; // 'user' | 'program'
	}

	export interface System {
		id: string; // Kept from previous logic, maps to top-level 'name' in new JSON
		name: string; // This is the ID like "3971b52c..."
		hostname: string;
		datetime: string;
		distribution: Distribution;
		fingerprint: Fingerprint;
		bios: BiosInfo | null;
		board: BoardInfo | null;
		devices: Device[] | null;
		drives: any | null; // Define if structure known
		incidents: any | null; // Define if structure known
		mitigations: any[] | null;
		networks: NetworkInterface[]; // Was NetworkInterface[] | null
		packages: PackageInfo[] | null; // For the 'packages' field, if used
		programs: any | null; // Define if structure known
		responses?: any[]; // Optional
		services: Service[] | null;
		antiques?: any[]; // Optional
		updates: PackageInfo[] | null; // For the 'updates' array
		users?: User[] | null;
	}

	let systems: System[] = [];
	let selectedSystem: System | null = null;
	let isLoading: boolean = true;
	let error: string | null = null;

	let panzoomInstance: PanzoomObject | null = null;
	let mapContainerElement: HTMLElement;
	let svgElement: SVGElement;

	// Make hexagons larger and adjust padding
	const hexSize = 160; // Increased from 120
	const hexWidth = hexSize;
	const hexHeight = Math.sqrt(3) * (hexSize / 2);
	const paddingX = hexWidth * 0.25; // Adjusted padding for larger size
	const paddingY = hexHeight * 0.20; // Adjusted padding
	const itemsPerRow = 4; // Might need to reduce items per row with larger hexagons

	onMount(async () => {
		try {
			isLoading = true;
			const listResponse = await fetch('/api/systems');
			if (!listResponse.ok) {
				throw new Error(`Failed to fetch system IDs: ${listResponse.statusText} (${listResponse.status})`);
			}
			const systemIds: string[] = await listResponse.json();

			const systemPromises = systemIds.map(async (id) => {
				const sysResponse = await fetch(`/api/systems/${id}`);
				if (!sysResponse.ok) {
					console.error(`Failed to fetch data for system ${id}: ${sysResponse.statusText} (${sysResponse.status})`);
					return null;
				}
				// The raw data from API
				const systemDataFromApi = await sysResponse.json();
				// Adapt to our System interface, especially mapping top-level 'name' to 'id' for consistency
				// and ensuring all fields match the new System interface.
				const adaptedSystem: System = {
					...systemDataFromApi, // Spread all fields from API
					id: systemDataFromApi.name, // map the unique name (like "3971b52c...") to 'id'
					// Ensure nested structures like 'networks' also conform if necessary
					// Example: if networks didn't have subnets directly but needed transformation
				};
				return adaptedSystem;
			});

			const resolvedSystems = await Promise.all(systemPromises);
			systems = resolvedSystems.filter(s => s !== null) as System[];
			error = null;

			if (svgElement && systems.length > 0) {
				panzoomInstance = Panzoom(svgElement, {
					maxZoom: 5, // Increased maxZoom
					minZoom: 0.2, // Decreased minZoom
					contain: 'outside',
					canvas: true,
					// cursor: 'grab', // Default panzoom cursor
					// panOnlyWhenZoomed: true,
					// pinchSpeed: 1,
					// initialZoom: 0.8, // Optionally set an initial zoom
					// initialX: (svgElement.parentElement?.clientWidth || 0) / 2 - svgContentWidth / 2,
					// initialY: (svgElement.parentElement?.clientHeight || 0) / 2 - svgContentHeight / 2,
				});
				// Attempt to center the content initially
				// This might need adjustment based on actual SVG content size vs container size
				setTimeout(() => {
					if (panzoomInstance && mapContainerElement && svgContentWidth > 0 && svgContentHeight > 0) {
						const containerWidth = mapContainerElement.clientWidth;
						const containerHeight = mapContainerElement.clientHeight;
						const initialScale = Math.min(containerWidth / svgContentWidth, containerHeight / svgContentHeight, 1);
						
						panzoomInstance.zoomTo(
							containerWidth / 2, // Center of the container
							containerHeight / 2,
							initialScale * 0.9 // Apply 90% of calculated scale
						);
						// panzoomInstance.moveTo(
						// 	(containerWidth - svgContentWidth * initialScale) / 2,
						// 	(containerHeight - svgContentHeight * initialScale) / 2
						// );
					}
				}, 100);
			}

		} catch (e: any) {
			console.error("Error in onMount:", e);
			error = e.message || 'An unknown error occurred while loading system data.';
			systems = [];
		} finally {
			isLoading = false;
		}
	});

	onDestroy(() => {
		if (panzoomInstance) {
			panzoomInstance.dispose();
		}
	});

	function selectSystem(system: System) {
		selectedSystem = system;
	}


    // Calculate total SVG dimensions needed for the grid
    $: totalRows = Math.max(1, Math.ceil((systems?.length || 0) / itemsPerRow));
    $: svgContentWidth = Math.max(100, itemsPerRow * (hexWidth + paddingX) - paddingX + hexWidth/2); // Add some buffer for staggered rows
    $: svgContentHeight = Math.max(100, totalRows * (hexHeight + paddingY) - paddingY + hexHeight/2);

</script>

<div class="min-h-screen h-screen bg-slate-100 text-slate-800 flex flex-col font-sans">
	{#if isLoading}
		<p class="text-center py-10 text-lg font-semibold text-slate-600">Loading Systems & Map...</p>
	{:else if error}
		<div class="p-6">
			<p class="text-red-600 font-semibold text-lg">Error Loading Systems:</p>
			<p class="text-red-500 bg-red-50 p-3 rounded mt-2">{error}</p>
			<p class="mt-2 text-sm text-slate-600">Please ensure the backend API is running at <code>http://localhost:3000</code> (proxied via <code>/api</code>) and returning valid data.</p>
		</div>
	{:else}
		<h1 class="text-3xl font-bold p-4 pb-2 text-slate-700 shadow-sm bg-white">Cybersecurity Hexagon Map</h1>
		<div class="flex-grow flex overflow-hidden"> 
			<!-- Hexagon map area with Panzoom -->
			<div bind:this={mapContainerElement} class="w-3/4 bg-slate-200 relative overflow-hidden select-none border-r border-slate-300">
				{#if systems.length > 0}
					<svg 
						bind:this={svgElement} 
						class="block" 
						style="width: {svgContentWidth}px; height: {svgContentHeight}px; transform-origin: 0 0;"
						preserveAspectRatio="xMidYMid meet" 
						><!-- viewBox is managed by panzoom -->
						<g id="hexagon-layer"> 
							{#each systems as system, i (system.id)}
								{@const col = i % itemsPerRow}
								{@const row = Math.floor(i / itemsPerRow)}
								{@const x = col * (hexWidth + paddingX) + hexWidth / 2 + (row % 2 === 1 ? (hexWidth + paddingX) / 2 : 0)}
								{@const y = row * (hexHeight * 0.75 + paddingY) + hexHeight / 2}

								<g transform="translate({x}, {y})" on:click={() => selectSystem(system)} class="cursor-pointer">
									<Hexagon {system} size={hexSize} selected={selectedSystem?.id === system.id} />
								</g>
							{/each}
						</g>
					</svg>
				{:else}
					<p class="text-center text-slate-500 p-10">No systems found to display on the map.</p>
				{/if}
			</div>

			<!-- Sidebar for selected system -->
			{#if selectedSystem}
				<div class="w-1/4 p-4 bg-white shadow-lg overflow-y-auto text-sm">
					<div class="flex justify-between items-center mb-3 pb-2 border-b">
						<h2 class="text-xl font-semibold text-slate-700">{selectedSystem.hostname}</h2>
						<button on:click={() => selectedSystem = null} class="text-slate-500 hover:text-red-600 p-1 rounded hover:bg-red-100 transition-colors">
							<i class="fas fa-times fa-lg"></i>
						</button>
					</div>
					
					<div class="space-y-2">
						<div><strong>ID:</strong> <span class="text-slate-600">{selectedSystem.id}</span></div>
						<div><strong>OS:</strong> <span class="text-slate-600">{selectedSystem.distribution.name} {selectedSystem.distribution.version}</span></div>
						<div><strong>Kernel:</strong> <span class="text-slate-600">{selectedSystem.distribution.kernel_version || selectedSystem.distribution.kernel}</span></div>
						<div><strong>Timezone:</strong> <span class="text-slate-600">{selectedSystem.fingerprint.timezone}</span></div>
						
						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">Services ({selectedSystem.services?.length || 0}):</h3>
							{#if selectedSystem.services && selectedSystem.services.length > 0}
								<div class="space-y-3">
									{#each selectedSystem.services as serviceItem}
										<div class="p-2 border rounded bg-slate-50">
											<p class="font-medium text-slate-800">{serviceItem.name} <span class="text-xs text-slate-500">(PID: {serviceItem.pid})</span></p>
											<p class="text-xs mt-0.5"><strong>Cmd:</strong> <code class="text-xs bg-slate-200 px-1 rounded">{serviceItem.command}</code></p>
											<p class="text-xs"><strong>User:</strong> {serviceItem.user.name} ({serviceItem.user.type})</p>
											{#if serviceItem.connections && serviceItem.connections.length > 0}
												<p class="text-xs mt-1"><strong>Listening:</strong></p>
												<ul class="list-disc list-inside pl-3 text-xs">
													{#each serviceItem.connections as conn}
														<li>{conn.protocol.toUpperCase()} on {conn.socket.host}:{conn.socket.port} ({conn.type})</li>
													{/each}
												</ul>
											{/if}
											<!-- Can add toggles for environment, filesystem etc. later -->
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-xs text-slate-500">No specific services listed for this system.</p>
							{/if}
						</div>

						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">Network Interfaces:</h3>
							{#if selectedSystem.networks && selectedSystem.networks.length > 0}
								{#each selectedSystem.networks as net}
									<div class="mb-2 p-2 border rounded bg-slate-50">
										<div><strong>IF:</strong> {net.name}</div>
										{#if net.subnets && net.subnets.length > 0}
											<ul class="list-disc list-inside pl-3 text-xs">
											{#each net.subnets as subnet}
												<li>{subnet.address}/{subnet.prefix} <span class="text-slate-500">({subnet.type}, {subnet.scope})</span></li>
											{/each}
											</ul>
										{/if}
									</div>
								{/each}
							{:else}
								<p class="text-xs text-slate-500">No network interfaces listed.</p>
							{/if}
						</div>

						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">Software & Updates ({selectedSystem.updates?.length || 0}):</h3>
							{#if selectedSystem.updates && selectedSystem.updates.length > 0}
								<ul class="list-disc list-inside max-h-48 overflow-y-auto text-xs space-y-0.5 pl-1">
									{#each selectedSystem.updates as pkg}
										<li title="Version: {pkg.version}
Manager: {pkg.manager}
Architecture: {pkg.architecture}
URL: {pkg.url || 'N/A'}">{pkg.name}</li>
									{/each}
								</ul>
							{:else}
								<p class="text-xs text-slate-500">No software updates or packages listed in 'updates' field.</p>
							{/if}
							{#if selectedSystem.packages && selectedSystem.packages.length > 0}
								<h4 class="font-medium mt-2 mb-1 text-xs text-slate-600">Other Packages ({selectedSystem.packages.length}):</h4>
								<ul class="list-disc list-inside max-h-32 overflow-y-auto text-xs space-y-0.5 pl-1">
									{#each selectedSystem.packages as pkg}
										<li>{pkg.name} ({pkg.version})</li>
									{/each}
								</ul>
							{/if}
						</div>

						{#if selectedSystem.devices && selectedSystem.devices.length > 0}
						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">Devices:</h3>
							<ul class="list-disc list-inside max-h-48 overflow-y-auto text-xs space-y-0.5 pl-1">
								{#each selectedSystem.devices as device}
									<li>{device.name} <span class="text-slate-500">({device.bus}, {device.system.vendor}:{device.system.device})</span></li>
								{/each}
							</ul>
						</div>
						{/if}
						
						<!-- Placeholder for Users, can be expanded -->
						{#if selectedSystem.users && selectedSystem.users.length > 0}
						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">Users ({selectedSystem.users.length}):</h3>
							<p class="text-xs text-slate-500">(Details can be expanded here)</p>
						</div>
						{/if}

					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="postcss">
	@tailwind base;
	@tailwind components;
	@tailwind utilities;

	/* Panzoom may add its own cursor styles; ensure compatibility or override if needed */
	.panzoom-grabbing {
		cursor: grabbing;
	}
</style>
