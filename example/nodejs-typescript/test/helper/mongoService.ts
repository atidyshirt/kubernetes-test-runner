import { MongoClient, Db, Collection } from 'mongodb';

export class MongoService {
  private client: MongoClient;
  private db!: Db;
  private collection!: Collection;

  constructor(connectionString: string = 'mongodb://mongodb:27017') {
    this.client = new MongoClient(connectionString);
  }

  async connect(): Promise<void> {
    try {
      await this.client.connect();
      this.db = this.client.db('testdb');
      this.collection = this.db.collection('testdata');
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

  async insertData(name: string, value: string): Promise<any> {
    try {
      const result = await this.collection.insertOne({
        name,
        value,
        timestamp: new Date()
      });
      return result;
    } catch (error) {
      throw new Error(`Failed to insert data: ${error}`);
    }
  }

  async getData(name: string): Promise<any> {
    try {
      return await this.collection.findOne({ name });
    } catch (error) {
      throw new Error(`Failed to get data: ${error}`);
    }
  }

  async clearData(): Promise<void> {
    try {
      await this.collection.deleteMany({});
    } catch (error) {
      console.warn('Warning: Failed to clear data:', error);
    }
  }

  async countDocuments(): Promise<number> {
    try {
      return await this.collection.countDocuments();
    } catch (error) {
      throw new Error(`Failed to count documents: ${error}`);
    }
  }
}
