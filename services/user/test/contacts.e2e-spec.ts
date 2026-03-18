import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication, ValidationPipe } from '@nestjs/common';
import request from 'supertest';
import { App } from 'supertest/types';
import { JwtService } from '@nestjs/jwt';
import { AppModule } from '../src/app.module';
import { PrismaService } from '../src/prisma/prisma.service';
import cookieParser from 'cookie-parser';

describe('Contacts (e2e)', () => {
  let app: INestApplication<App>;
  let accessToken: string;

  const mockPrisma = {
    user: {
      create: jest.fn(),
      findUnique: jest.fn(),
      findMany: jest.fn(),
      count: jest.fn(),
      update: jest.fn(),
      delete: jest.fn(),
    },
    refreshToken: {
      create: jest.fn(),
      findUnique: jest.fn(),
      delete: jest.fn(),
      deleteMany: jest.fn(),
    },
    contact: {
      create: jest.fn(),
      findMany: jest.fn(),
      findUnique: jest.fn(),
      update: jest.fn(),
      delete: jest.fn(),
      deleteMany: jest.fn(),
      count: jest.fn(),
    },
    blockedUser: {
      create: jest.fn(),
      findFirst: jest.fn(),
      findMany: jest.fn(),
      findUnique: jest.fn(),
      delete: jest.fn(),
      deleteMany: jest.fn(),
      count: jest.fn(),
    },
    $transaction: jest.fn(),
    $connect: jest.fn(),
    $disconnect: jest.fn(),
  };

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    })
      .overrideProvider(PrismaService)
      .useValue(mockPrisma)
      .compile();

    app = moduleFixture.createNestApplication();
    app.use(cookieParser());
    app.useGlobalPipes(new ValidationPipe({ whitelist: true, transform: true }));
    await app.init();

    const jwtService = moduleFixture.get<JwtService>(JwtService);
    accessToken = jwtService.sign({
      sub: 'user-uuid',
      username: 'testuser',
      role: 'user',
    });
  });

  afterAll(async () => {
    await app.close();
  });

  beforeEach(() => {
    jest.clearAllMocks();
  });

  // ========================
  // CONTACTS
  // ========================

  describe('POST /contacts', () => {
    it('should add a contact', async () => {
      const contactUuid = '550e8400-e29b-41d4-a716-446655440000';
      mockPrisma.user.findUnique.mockResolvedValue({ id: contactUuid });
      mockPrisma.blockedUser.findFirst.mockResolvedValue(null);
      mockPrisma.contact.create.mockResolvedValue({
        id: 'entry-uuid',
        userId: 'user-uuid',
        contactUserId: contactUuid,
        nickname: null,
        contactUser: { id: contactUuid, username: 'friend' },
      });

      const res = await request(app.getHttpServer())
        .post('/contacts')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ contactUserId: contactUuid })
        .expect(201);

      expect(res.body).toHaveProperty('contactUserId', contactUuid);
    });

    it('should return 400 for invalid UUID format', async () => {
      await request(app.getHttpServer())
        .post('/contacts')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ contactUserId: 'not-a-uuid' })
        .expect(400);
    });

    it('should return 400 for missing contactUserId', async () => {
      await request(app.getHttpServer())
        .post('/contacts')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({})
        .expect(400);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .post('/contacts')
        .send({ contactUserId: '550e8400-e29b-41d4-a716-446655440000' })
        .expect(401);
    });
  });

  describe('GET /contacts', () => {
    it('should return user contacts', async () => {
      mockPrisma.contact.findMany.mockResolvedValue([]);
      mockPrisma.contact.count.mockResolvedValue(0);

      const res = await request(app.getHttpServer())
        .get('/contacts')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toHaveProperty('data');
      expect(res.body).toHaveProperty('total', 0);
      expect(res.body).toHaveProperty('page', 1);
      expect(res.body).toHaveProperty('limit', 20);
    });

    it('should accept page and limit params', async () => {
      mockPrisma.contact.findMany.mockResolvedValue([]);
      mockPrisma.contact.count.mockResolvedValue(0);

      const res = await request(app.getHttpServer())
        .get('/contacts?page=2&limit=5')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toHaveProperty('page', 2);
      expect(res.body).toHaveProperty('limit', 5);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .get('/contacts')
        .expect(401);
    });
  });

  describe('PATCH /contacts/:id', () => {
    it('should update contact nickname', async () => {
      const contactId = '550e8400-e29b-41d4-a716-446655440000';
      mockPrisma.contact.findUnique.mockResolvedValue({
        id: contactId,
        userId: 'user-uuid',
      });
      mockPrisma.contact.update.mockResolvedValue({
        id: contactId,
        userId: 'user-uuid',
        nickname: 'New Name',
        contactUser: { id: 'friend-uuid', username: 'friend' },
      });

      const res = await request(app.getHttpServer())
        .patch(`/contacts/${contactId}`)
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ nickname: 'New Name' })
        .expect(200);

      expect(res.body).toHaveProperty('nickname', 'New Name');
    });

    it('should return 400 for invalid UUID param', async () => {
      await request(app.getHttpServer())
        .patch('/contacts/not-a-uuid')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ nickname: 'test' })
        .expect(400);
    });

    it('should return 404 if contact not found or not owned', async () => {
      const contactId = '550e8400-e29b-41d4-a716-446655440000';
      mockPrisma.contact.findUnique.mockResolvedValue(null);

      await request(app.getHttpServer())
        .patch(`/contacts/${contactId}`)
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ nickname: 'test' })
        .expect(404);
    });
  });

  describe('DELETE /contacts/:id', () => {
    it('should remove a contact', async () => {
      const contactId = '550e8400-e29b-41d4-a716-446655440000';
      mockPrisma.contact.findUnique.mockResolvedValue({
        id: contactId,
        userId: 'user-uuid',
      });
      mockPrisma.contact.delete.mockResolvedValue({});

      const res = await request(app.getHttpServer())
        .delete(`/contacts/${contactId}`)
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toEqual({ deleted: true });
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .delete('/contacts/550e8400-e29b-41d4-a716-446655440000')
        .expect(401);
    });
  });

  // ========================
  // BLOCKED USERS
  // ========================

  describe('POST /contacts/blocked', () => {
    it('should block a user', async () => {
      const blockedUuid = '550e8400-e29b-41d4-a716-446655440001';
      mockPrisma.user.findUnique.mockResolvedValue({ id: blockedUuid });
      const blocked = {
        id: 'blocked-entry-uuid',
        userId: 'user-uuid',
        blockedUserId: blockedUuid,
        blockedUser: { id: blockedUuid, username: 'baduser' },
      };
      mockPrisma.$transaction.mockResolvedValue([blocked, { count: 0 }]);

      const res = await request(app.getHttpServer())
        .post('/contacts/blocked')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ blockedUserId: blockedUuid })
        .expect(201);

      expect(res.body).toHaveProperty('blockedUserId', blockedUuid);
    });

    it('should return 400 for invalid UUID', async () => {
      await request(app.getHttpServer())
        .post('/contacts/blocked')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ blockedUserId: 'invalid' })
        .expect(400);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .post('/contacts/blocked')
        .send({ blockedUserId: '550e8400-e29b-41d4-a716-446655440001' })
        .expect(401);
    });
  });

  describe('GET /contacts/blocked', () => {
    it('should return blocked users', async () => {
      mockPrisma.blockedUser.findMany.mockResolvedValue([]);
      mockPrisma.blockedUser.count.mockResolvedValue(0);

      const res = await request(app.getHttpServer())
        .get('/contacts/blocked')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toHaveProperty('data');
      expect(res.body).toHaveProperty('total', 0);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .get('/contacts/blocked')
        .expect(401);
    });
  });

  describe('DELETE /contacts/blocked/:id', () => {
    it('should unblock a user', async () => {
      const entryId = '550e8400-e29b-41d4-a716-446655440000';
      mockPrisma.blockedUser.findUnique.mockResolvedValue({
        id: entryId,
        userId: 'user-uuid',
        blockedUserId: 'blocked-uuid',
      });
      mockPrisma.blockedUser.delete.mockResolvedValue({});

      const res = await request(app.getHttpServer())
        .delete(`/contacts/blocked/${entryId}`)
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toEqual({ deleted: true });
    });

    it('should return 404 if entry not found', async () => {
      const entryId = '550e8400-e29b-41d4-a716-446655440000';
      mockPrisma.blockedUser.findUnique.mockResolvedValue(null);

      await request(app.getHttpServer())
        .delete(`/contacts/blocked/${entryId}`)
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(404);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .delete('/contacts/blocked/550e8400-e29b-41d4-a716-446655440000')
        .expect(401);
    });
  });
});
