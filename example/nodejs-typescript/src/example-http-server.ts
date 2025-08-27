import express from 'express';
import { MongoClient } from 'mongodb';

const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

let mongoClient: MongoClient;
let db: any;

async function connectToMongo() {
  try {
    mongoClient = new MongoClient('mongodb://mongodb:27017');
    await mongoClient.connect();
    db = mongoClient.db('testdb');
  } catch (error) {
    console.error('Failed to connect to MongoDB:', error);
  }
}

app.post('/data', async (req, res) => {
  try {
    const { name, value } = req.body;
    if (!name || !value) {
      return res.status(400).json({ error: 'Name and value are required' });
    }

    const result = await db.collection('testdata').insertOne({
      name,
      value,
      timestamp: new Date()
    });

    res.json({ 
      success: true, 
      id: result.insertedId,
      message: 'Data inserted successfully'
    });
  } catch (error) {
    console.error('Error inserting data:', error);
    res.status(500).json({ error: 'Failed to insert data' });
  }
});

app.get('/data/:name', async (req, res) => {
  try {
    const { name } = req.params;
    const data = await db.collection('testdata').findOne({ name });
    
    if (!data) {
      return res.status(404).json({ error: 'Data not found' });
    }

    res.json(data);
  } catch (error) {
    console.error('Error retrieving data:', error);
    res.status(500).json({ error: 'Failed to retrieve data' });
  }
});

app.get('/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date() });
});

async function startServer() {
  await connectToMongo();
  
  app.listen(port, () => {
    console.log(`Server running on port ${port}`);
  });
}

if (require.main === module) {
  startServer();
}

export { app, startServer };
