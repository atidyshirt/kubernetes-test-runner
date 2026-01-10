export class WiremockService {
    private namespace: string;

    constructor(namespace: string) {
        this.namespace = namespace;
    }

    private adminMappingsUrl(serviceName: string): string {
        return `http://${serviceName}.${this.namespace}.svc.cluster.local:3000/__admin/mappings`;
    }

    private async getAllMappings(serviceName: string) {
        const response = await fetch(this.adminMappingsUrl(serviceName));
        if (!response.ok) {
            throw new Error(`Failed to fetch mappings: ${await response.text()}`);
        }
        const { mappings } = (await response.json()) as any;
        return mappings;
    }

    private async findMappingUuidByRequest(
        serviceName: string,
        endpoint: string,
        method: string,
    ): Promise<string | null> {
        const mappings = await this.getAllMappings(serviceName);
        const mapping = mappings.find(
            (m: any) => m.request.url === endpoint && m.request.method === method,
        );
        return mapping !== undefined ? mapping.uuid : null;
    }

    async updateMapping(
        serviceName: string,
        endpoint: string,
        method: string,
        jsonBody: object,
    ): Promise<void> {
        const uuid = await this.findMappingUuidByRequest(serviceName, endpoint, method);
        if (uuid === null) {
            throw new Error('Mapping not found');
        }
        const url = `${this.adminMappingsUrl(serviceName)}/${uuid}`;
        const updatedMapping = {
            request: {
                url: endpoint,
                method,
            },
            response: {
                status: 200,
                jsonBody,
                headers: { 'Content-Type': 'application/json' },
            },
        };
        const response = await fetch(url, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updatedMapping),
        });
        if (!response.ok) {
            throw new Error(`Failed to update mapping: ${await response.text()}`);
        }
    }

    async resetMappings(serviceName: string): Promise<void> {
        const url = `http://${serviceName}.${this.namespace}.svc.cluster.local:3000/__admin/mappings/reset`;
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
        });
        if (!response.ok) {
            throw new Error(
                `Failed to reset mappings for ${serviceName}: ${await response.text()}`,
            );
        }
    }
}
