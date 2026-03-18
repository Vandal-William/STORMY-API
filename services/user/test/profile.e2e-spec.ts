import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication, ValidationPipe } from '@nestjs/common';
import request from 'supertest';
import { App } from 'supertest/types';
import { JwtService } from '@nestjs/jwt';
import { AppModule } from '../src/app.module';
import { PrismaService } from '../src/prisma/prisma.service';
import cookieParser from 'cookie-parser';

describe('Profile (e2e)', () => {
  let app: INestApplication<App>;
  let jwtService: JwtService;
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
    app.useGlobalPipes(
      new ValidationPipe({ whitelist: true, transform: true }),
    );
    await app.init();

    jwtService = moduleFixture.get<JwtService>(JwtService);
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

  describe('GET /profile/me', () => {
    it('should return own profile when authenticated', async () => {
      const profile = {
        id: 'user-uuid',
        phone: '0612345678',
        username: 'testuser',
        email: 'test@example.com',
        avatarUrl: null,
        about: null,
        lastSeen: new Date().toISOString(),
        createdAt: new Date().toISOString(),
      };
      mockPrisma.user.findUnique.mockResolvedValue(profile);

      const res = await request(app.getHttpServer())
        .get('/profile/me')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toHaveProperty('id', 'user-uuid');
      expect(res.body).toHaveProperty('username', 'testuser');
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .get('/profile/me')
        .expect(401);
    });
  });

  describe('GET /profile/search', () => {
    it('should search users by username', async () => {
      const users = [
        {
          id: '1',
          username: 'alice',
          avatarUrl: null,
          about: null,
          lastSeen: new Date(),
        },
      ];
      mockPrisma.user.findMany.mockResolvedValue(users);
      mockPrisma.user.count.mockResolvedValue(1);

      const res = await request(app.getHttpServer())
        .get('/profile/search?username=alice')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toHaveProperty('data');
      expect(res.body).toHaveProperty('total', 1);
    });

    it('should return 400 if username query is missing', async () => {
      await request(app.getHttpServer())
        .get('/profile/search')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(400);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .get('/profile/search?username=alice')
        .expect(401);
    });
  });

  describe('GET /profile/:id', () => {
    it('should return public profile of a user', async () => {
      const user = {
        id: 'other-uuid',
        username: 'otheruser',
        avatarUrl: null,
        about: 'Bio',
        lastSeen: new Date(),
      };
      mockPrisma.user.findUnique.mockResolvedValue(user);

      const res = await request(app.getHttpServer())
        .get('/profile/550e8400-e29b-41d4-a716-446655440000')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toHaveProperty('username');
    });

    it('should return 400 for invalid UUID', async () => {
      await request(app.getHttpServer())
        .get('/profile/not-a-uuid')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(400);
    });

    it('should return 404 if user not found', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await request(app.getHttpServer())
        .get('/profile/550e8400-e29b-41d4-a716-446655440000')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(404);
    });
  });

  describe('PATCH /profile/me', () => {
    it('should update own profile', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'testuser',
      });
      mockPrisma.user.update.mockResolvedValue({
        id: 'user-uuid',
        phone: '0612345678',
        username: 'newname',
        email: 'test@example.com',
        avatarUrl: null,
        about: null,
        lastSeen: new Date(),
        createdAt: new Date(),
      });

      const res = await request(app.getHttpServer())
        .patch('/profile/me')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ username: 'newname' })
        .expect(200);

      expect(res.body).toHaveProperty('username', 'newname');
    });

    it('should return 400 for invalid data (username too short)', async () => {
      await request(app.getHttpServer())
        .patch('/profile/me')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .send({ username: 'ab' })
        .expect(400);
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .patch('/profile/me')
        .send({ username: 'newname' })
        .expect(401);
    });
  });

  describe('DELETE /profile/me', () => {
    it('should delete own account', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'testuser',
      });
      mockPrisma.$transaction.mockResolvedValue([
        { count: 0 },
        { count: 0 },
        { id: 'user-uuid' },
      ]);

      const res = await request(app.getHttpServer())
        .delete('/profile/me')
        .set('Cookie', [`ACCESS_TOKEN=${accessToken}`])
        .expect(200);

      expect(res.body).toEqual({ deleted: true });
    });

    it('should return 401 without auth', async () => {
      await request(app.getHttpServer())
        .delete('/profile/me')
        .expect(401);
    });
  });
});
