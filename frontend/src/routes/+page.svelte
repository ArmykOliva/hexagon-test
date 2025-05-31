<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Panzoom from 'panzoom'; // Import Panzoom
	import type { PanzoomObject } from 'panzoom'; // Import PanzoomObject type
	import Hexagon from '../lib/components/Hexagon.svelte'; // Import Hexagon component
	import type { 
		System, 
		// Service, // Service is used within System, direct import not strictly needed here if only selectedSystem.services is used
		// NetworkInterface, // Same, used within System
		// Ensure all types used directly in the template or script logic (not just nested in System) are imported.
		SystemToSystemConnection, // Import for use in template
		SystemConnectionSocket, // Ensure these are imported for mock data
		User, UserGroup // Import for use in template
	} from '../lib/types'; // Updated import path

	// --- TypeScript Interfaces are now in ../lib/types.ts ---

	// --- Mocked System-to-System Connections ---
	// Using the IDs you provided for OlivaSurface and thinky
	const MOCKED_CONNECTIONS_CONFIG: Array<{sourceId: string, targetId: string, protocol?: string, type?: string, sourcePort?: number, targetPort?: number}> = [
		{
			sourceId: "3971b52ceebf43e7b960010b2f976e30", // Assuming this is OlivaSurface ID
			targetId: "6e5b93f546924feca5fd07c681b92f80", // Assuming this is Thinky ID
			protocol: "tcp", 
			type: "simulated-link",
			sourcePort: 80, // Example port
			targetPort: 54321 // Example port
		}
		// Add more mock connections here if needed
	];

	let systems: System[] = [];
	let selectedSystem: System | null = null;
	let isLoading: boolean = true;
	let error: string | null = null;

	let panzoomInstance: PanzoomObject | null = null;
	let mapContainerElement: HTMLElement;
	let svgElement: SVGElement;

	// Store hexagon positions for drawing lines
	let hexagonPositions: Record<string, { x: number; y: number; q: number; r: number }> = {};

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
				const systemDataFromApi = await sysResponse.json();
				const baseSystem: System = {
					...systemDataFromApi,
					id: systemDataFromApi.name, 
					connections: systemDataFromApi.connections || [] // Ensure connections array exists
				};
				return baseSystem;
			});

			let resolvedSystems = (await Promise.all(systemPromises)).filter(s => s !== null) as System[];

			// Augment systems with mocked connections for sidebar display and layout hinting
			resolvedSystems = resolvedSystems.map(sys => {
				const outgoingMockConnections: SystemToSystemConnection[] = [];
				MOCKED_CONNECTIONS_CONFIG.forEach(mockConf => {
					if (mockConf.sourceId === sys.id) {
						outgoingMockConnections.push({
							source: { host: mockConf.sourceId, port: mockConf.sourcePort || 0 },
							target: { host: mockConf.targetId, port: mockConf.targetPort || 0 },
							protocol: mockConf.protocol || "unknown",
							type: mockConf.type || "mocked"
						});
					}
					// Also add reverse connection if you want two-way visibility from one-way config
					// For simplicity, current mock config is one-way from sourceId
				});
				return {
					...sys,
					connections: [...(sys.connections || []), ...outgoingMockConnections]
				};
			});

			systems = resolvedSystems;
			error = null;

			// --- Connection-Aware Hexagon Layout (Revised) --- 
			const newHexagonPositions: Record<string, { x: number; y: number; q: number; r: number }> = {};
			const occupiedAxialCoords: Set<string> = new Set(); // "q,r"
			
			function axialToPixel(q: number, r: number): { x: number; y: number } {
				// For pointy-topped hexagons, a common conversion:
				// x = hex_radius * (sqrt(3) * q + sqrt(3)/2 * r)
				// y = hex_radius * (3/2 * r)
				// Using our hexSize (diameter) and padding:
				const x = (hexWidth + paddingX) * q + (hexWidth + paddingX) / 2 * r; // Adjusted for axial coordinates
        		const y = (hexHeight * 0.75 + paddingY) * r;
				return { x: x + hexWidth / 2, y: y + hexHeight / 2 }; 
			}

			// Directions for axial coordinates (pointy-topped)
			const axialDirections = [
				{q: +1, r:  0}, {q: +1, r: -1}, {q:  0, r: -1}, 
    			{q: -1, r:  0}, {q: -1, r: +1}, {q:  0, r: +1}
			];

			const systemsToPlaceInitially = [...systems];
			const placedSystemIds = new Set<string>();

			// IDs for your specific mocked connection
			const OLIVA_ID = "3971b52ceebf43e7b960010b2f976e30";
			const THINKY_ID = "6e5b93f546924feca5fd07c681b92f80";

			// 1. Place OlivaSurface (if it exists)
			const olivaSystemIndex = systemsToPlaceInitially.findIndex(s => s.id === OLIVA_ID);
			if (olivaSystemIndex !== -1) {
				const olivaSystem = systemsToPlaceInitially.splice(olivaSystemIndex, 1)[0];
				const q = 0, r = 0;
				const pixelPos = axialToPixel(q, r);
				newHexagonPositions[olivaSystem.id] = { ...pixelPos, q, r };
				occupiedAxialCoords.add(`${q},${r}`);
				placedSystemIds.add(olivaSystem.id);
				console.log(`Placed ${olivaSystem.hostname} (ID: ${olivaSystem.id}) at q:${q}, r:${r}`);
			}

			// 2. Place Thinky (if it exists and Oliva was placed) adjacently
			const thinkySystemIndex = systemsToPlaceInitially.findIndex(s => s.id === THINKY_ID);
			if (thinkySystemIndex !== -1 && newHexagonPositions[OLIVA_ID]) {
				const thinkySystem = systemsToPlaceInitially.splice(thinkySystemIndex, 1)[0];
				const olivaCoords = newHexagonPositions[OLIVA_ID];
				let placedThinky = false;
				for (const dir of axialDirections) {
					const nq = olivaCoords.q + dir.q;
					const nr = olivaCoords.r + dir.r;
					if (!occupiedAxialCoords.has(`${nq},${nr}`)) {
						const pixelPos = axialToPixel(nq, nr);
						newHexagonPositions[thinkySystem.id] = { ...pixelPos, q: nq, r: nr };
						occupiedAxialCoords.add(`${nq},${nr}`);
						placedSystemIds.add(thinkySystem.id);
						console.log(`Placed ${thinkySystem.hostname} (ID: ${thinkySystem.id}) at q:${nq}, r:${nr} (neighbor to Oliva)`);
						placedThinky = true;
						break;
					}
				}
				if (!placedThinky) {
					console.warn(`Could not place ${thinkySystem.hostname} adjacently to Oliva. All neighbors occupied. Placing with spiral.`);
					systemsToPlaceInitially.unshift(thinkySystem); // Add back to be placed by spiral
				}
			}

			// 3. Place remaining systems using a hexagonal spiral
			let q = 0, r = 0;
			let leg = 0; // current leg of the spiral
			let stepInLeg = 0;
			let dirIndex = 0; // initial direction

			systemsToPlaceInitially.forEach(system => {
				if (placedSystemIds.has(system.id)) return; // Should have been handled if it was Oliva/Thinky

				// Find next free spot in spiral from (0,0)
				q=0; r=0; leg=0; stepInLeg=0; dirIndex=0;
				// Loop to find an actually free spot using spiral logic
				while(occupiedAxialCoords.has(`${q},${r}`)) {
					if (stepInLeg >= Math.floor(leg / 2) + 1 && leg !==0 ) { // For leg length definition that increases every 2 legs
					// if (stepInLeg >= legLengths[leg]) { // Alternative: precomputed leg lengths
						dirIndex = (dirIndex + 1) % 6;
						stepInLeg = 0;
						leg++; // This simple leg increment doesn't give standard spiral lengths. Needs fix for true spiral.
					}
					// A more robust spiral would recalculate leg based on ring number
					// This simplified 'leg++' will make a less compact spiral.
					// To ensure it moves, always take a step
					q += axialDirections[dirIndex].q;
					r += axialDirections[dirIndex].r;
					stepInLeg++;
					// Failsafe to prevent infinite loop if spiral logic is flawed for finding *next* step
					// This part of spiral needs to be more robust for general purpose filling.
					// The core issue with simple spiral is ensuring it tries *new* cells systematically.
				}

				const pixelPos = axialToPixel(q, r);
				newHexagonPositions[system.id] = { ...pixelPos, q, r };
				occupiedAxialCoords.add(`${q},${r}`);
				placedSystemIds.add(system.id);
				// console.log(`Placed ${system.hostname} (ID: ${system.id}) at q:${q}, r:${r} via spiral`);
			});

			hexagonPositions = newHexagonPositions;

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
								{@const pos = hexagonPositions[system.id] || {x:0, y:0}} 
								<g transform="translate({pos.x}, {pos.y})" on:click={() => selectSystem(system)} class="cursor-pointer">
									<Hexagon {system} size={hexSize} selected={selectedSystem?.id === system.id} />
								</g>
							{/each}
						</g>

						<!-- Add Connection Lines Layer -->
						<g id="connection-lines-layer" stroke="rgba(50, 50, 200, 0.6)" stroke-width="3">
							{#each systems as sourceSystem (sourceSystem.id)}
								{#if sourceSystem.connections && hexagonPositions[sourceSystem.id]}
									{#each sourceSystem.connections as conn}
										{@const targetSystemId = conn.target.host}
										{@const targetPos = hexagonPositions[targetSystemId]}
										{@const sourcePos = hexagonPositions[sourceSystem.id]}
										{#if targetPos && sourcePos}
											<line x1={sourcePos.x} y1={sourcePos.y} x2={targetPos.x} y2={targetPos.y} />
										{/if}
									{/each}
								{/if}
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
						
						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">Users ({selectedSystem.users?.length || 0}):</h3>
							{#if selectedSystem.users && selectedSystem.users.length > 0}
								<div class="space-y-2 max-h-48 overflow-y-auto">
									{#each selectedSystem.users as userItem (userItem.id)}
										<div class="p-1.5 border rounded bg-slate-50 text-xs">
											<p class="font-medium text-slate-800">{userItem.name} <span class="text-slate-500">(ID: {userItem.id}, Type: {userItem.type})</span></p>
											<p><strong>Shell:</strong> {userItem.shell || 'N/A'}</p>
											{#if userItem.groups && userItem.groups.length > 0}
												<p><strong>Groups:</strong> {userItem.groups.map(g => g.name).join(', ')}</p>
											{/if}
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-xs text-slate-500">No users listed for this system.</p>
							{/if}
						</div>

						<div class="pt-2 mt-2 border-t">
							<h3 class="text-md font-semibold mb-1 text-slate-700">System Connections ({selectedSystem.connections?.length || 0}):</h3>
							{#if selectedSystem.connections && selectedSystem.connections.length > 0}
								<div class="space-y-2 max-h-48 overflow-y-auto">
									{#each selectedSystem.connections as conn (conn.target.host + conn.target.port + conn.protocol)}
										{@const targetSystem = systems.find(s => s.id === conn.target.host)}
										<div class="p-1.5 border rounded bg-slate-50 text-xs">
											<p class="font-medium text-slate-800">
												Connects to: {targetSystem ? targetSystem.hostname : conn.target.host}
											</p>
											<p><strong>Via:</strong> {conn.target.port}/{conn.protocol} ({conn.type})</p>
											<p class="text-xs text-slate-500">Source: {conn.source.host}:{conn.source.port}</p>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-xs text-slate-500">No specific system connections listed.</p>
							{/if}
						</div>

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
