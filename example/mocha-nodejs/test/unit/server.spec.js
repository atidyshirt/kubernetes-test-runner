const { expect } = require('chai');
const request = require('supertest');
const app = require('../../src/server');

describe('Express Server Unit Tests', function() {
  this.timeout(10000);
  
  describe('Health Endpoint', function() {
    it('should respond to health check', async function() {
      const response = await request(app)
        .get('/health')
        .expect(200);
      
      expect(response.body).to.have.property('status', 'ok');
      expect(response.body).to.have.property('timestamp');
      expect(response.body).to.have.property('uptime');
      expect(response.body).to.have.property('environment');
    });
  });
  
  describe('Root Endpoint', function() {
    it('should return API information', async function() {
      const response = await request(app)
        .get('/')
        .expect(200);
      
      expect(response.body).to.have.property('message');
      expect(response.body.message).to.include('Hello from Express Server');
      expect(response.body).to.have.property('version');
      expect(response.body).to.have.property('endpoints');
    });
  });
  
  describe('Users API', function() {
    describe('GET /api/users', function() {
      it('should return all users', async function() {
        const response = await request(app)
          .get('/api/users')
          .expect(200);
        
        expect(response.body).to.be.an('array');
        expect(response.body).to.have.length.at.least(2);
        expect(response.body[0]).to.have.property('id');
        expect(response.body[0]).to.have.property('name');
        expect(response.body[0]).to.have.property('email');
      });
    });
    
    describe('GET /api/users/:id', function() {
      it('should return user by ID', async function() {
        const response = await request(app)
          .get('/api/users/1')
          .expect(200);
        
        expect(response.body).to.have.property('id', 1);
        expect(response.body).to.have.property('name');
        expect(response.body).to.have.property('email');
      });
      
      it('should return 404 for non-existent user', async function() {
        await request(app)
          .get('/api/users/999')
          .expect(404);
      });
    });
    
    describe('POST /api/users', function() {
      it('should create new user', async function() {
        const newUser = {
          name: 'Test User',
          email: 'test@example.com'
        };
        
        const response = await request(app)
          .post('/api/users')
          .send(newUser)
          .expect(201);
        
        expect(response.body).to.have.property('id');
        expect(response.body).to.have.property('name', newUser.name);
        expect(response.body).to.have.property('email', newUser.email);
        expect(response.body).to.have.property('createdAt');
      });
      
      it('should return 400 for missing fields', async function() {
        const invalidUser = { name: 'Test User' };
        
        await request(app)
          .post('/api/users')
          .send(invalidUser)
          .expect(400);
      });
      
      it('should return 409 for duplicate email', async function() {
        const duplicateUser = {
          name: 'Duplicate User',
          email: 'john@example.com' // Already exists
        };
        
        await request(app)
          .post('/api/users')
          .send(duplicateUser)
          .expect(409);
      });
    });
    
    describe('PUT /api/users/:id', function() {
      it('should update existing user', async function() {
        const updates = { name: 'Updated Name' };
        
        const response = await request(app)
          .put('/api/users/1')
          .send(updates)
          .expect(200);
        
        expect(response.body).to.have.property('name', updates.name);
        expect(response.body).to.have.property('updatedAt');
      });
      
      it('should return 404 for non-existent user', async function() {
        await request(app)
          .put('/api/users/999')
          .send({ name: 'Test' })
          .expect(404);
      });
      
      it('should return 400 for no update fields', async function() {
        await request(app)
          .put('/api/users/1')
          .send({})
          .expect(400);
      });
    });
    
    describe('DELETE /api/users/:id', function() {
      it('should delete existing user', async function() {
        const response = await request(app)
          .delete('/api/users/2')
          .expect(200);
        
        expect(response.body).to.have.property('message');
        expect(response.body).to.have.property('user');
        expect(response.body.user).to.have.property('id', 2);
      });
      
      it('should return 404 for non-existent user', async function() {
        await request(app)
          .delete('/api/users/999')
          .expect(404);
      });
    });
  });
  
  describe('Error Handling', function() {
    it('should return 404 for non-existent routes', async function() {
      await request(app)
        .get('/nonexistent')
        .expect(404);
    });
    
    it('should handle malformed JSON gracefully', async function() {
      await request(app)
        .post('/api/users')
        .set('Content-Type', 'application/json')
        .send('invalid json')
        .expect(400);
    });
  });
});
