import { MongoClient, Db } from 'mongodb';

export class MongoService {
    private client: MongoClient;
    private db!: Db;

    constructor(connectionString = 'mongodb://mongodb:27017') {
        this.client = new MongoClient(connectionString);
    }

    getDb(): Db {
        if (this.db === undefined) {
            throw new Error('MongoDB not connected');
        }
        return this.db;
    }

    async connect(): Promise<void> {
        try {
            await this.client.connect();
            this.db = this.client.db('testdb');
        } catch (error) {
            throw new Error(`Failed to connect to MongoDB: ${error}`);
        }
    }

    async disconnect(): Promise<void> {
        try {
            await this.client.close();
        } catch (error) {
            console.warn('Warning: Failed to disconnect from MongoDB:', error);
        }
    }
}
