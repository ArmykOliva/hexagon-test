<script lang="ts">
	// Import System and Service, but Package might not be needed if icons are from services only
	import type { System, Service } from '../../routes/+page.svelte'; 

	export let system: System;
	export let size: number = 100;
	export let selected: boolean = false;

	// Specific map for service names to Font Awesome icons
	const serviceIconMap: Record<string, string> = {
		'ssh': 'fas fa-terminal',
		'sshd': 'fas fa-terminal',
		'http': 'fas fa-server',
		'https': 'fas fa-server',
		'nginx': 'fas fa-server',
		'apache': 'fas fa-server',
		'apache2': 'fas fa-server',
		'mysql': 'fas fa-database',
		'mariadb': 'fas fa-database',
		'postgres': 'fas fa-database',
		'postgresql': 'fas fa-database',
		'ftp': 'fas fa-folder-open',
		'sftp': 'fas fa-folder-open',
		'dns': 'fas fa-network-wired', // or a more specific DNS icon if available
		'samba': 'fas fa-folder-tree',
		'smb': 'fas fa-folder-tree',
		'rdp': 'fas fa-desktop',
		'vnc': 'fas fa-desktop',
		'docker': 'fab fa-docker', // Note: Font Awesome brands require 'fab'
		// Add more service-to-icon mappings here
	};

	function getServiceIcon(serviceName: string): string {
		const nameLower = serviceName.toLowerCase();
		for (const key in serviceIconMap) {
			if (nameLower.includes(key)) {
				return serviceIconMap[key];
			}
		}
		return 'fas fa-cog'; // Default icon for unknown services
	}

	const r = size / 2;
	const hexHeight = Math.sqrt(3) * r;
	const points = [
		`${r},0`,
		`${r/2},${hexHeight/2}`,
		`${-r/2},${hexHeight/2}`,
		`${-r},0`,
		`${-r/2},-${hexHeight/2}`,
		`${r/2},-${hexHeight/2}`,
	].join(' ');

    // Determine service icons to display (max 3 for space, from system.services)
    let displayableServices: Service[] = [];
    if (system.services && system.services.length > 0) {
        // Simple slice for now, could be more sophisticated (e.g. prioritize critical services)
        displayableServices = system.services.slice(0, 3);
    }

</script>

<g class="group">
    <polygon 
        {points} 
        class="fill-blue-500 group-hover:fill-blue-600 transition-colors duration-150 {selected ? 'stroke-yellow-400' : 'stroke-white'}"
        stroke-width={selected ? "4" : "2"}
    />
    <text 
        x="0" 
        y="0" 
        text-anchor="middle" 
        dominant-baseline="middle" 
        class="fill-white font-semibold pointer-events-none select-none"
        style="font-size: {Math.max(8, size * 0.15)}px;"
    >
        {system.hostname}
    </text>

    <!-- Service Icons from system.services -->
    {#if displayableServices.length > 0}
    <g transform="translate(0, {hexHeight * 0.28})">
        {#each displayableServices as serviceItem, i}
            {@const iconClass = getServiceIcon(serviceItem.name)}
            {@const iconSize = size * 0.12}
            <foreignObject 
                x={ (i - (displayableServices.length -1) / 2) * (iconSize * 1.5) - iconSize / 2}
                y={-iconSize / 2}
                width={iconSize}
                height={iconSize}
                class="pointer-events-none"
            >
                <div xmlns="http://www.w3.org/1999/xhtml" class="flex items-center justify-center h-full">
                    <i class="{iconClass} {iconClass.startsWith('fab') ? '' : 'text-white'}" style="font-size: {iconSize * 0.8}px;" title={serviceItem.name}></i>
                </div>
            </foreignObject>
        {/each}
    </g>
    {/if}

    <!-- No need for separate selected polygon, combined with main polygon's stroke -->
</g>

<style>
    /* Ensure Font Awesome brands (fab) get their default color if not white */
    .fab {
        /* color: inherit; You might need to adjust if FA colors clash */
    }
</style> 