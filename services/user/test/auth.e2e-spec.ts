import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication, ValidationPipe } from '@nestjs/common';
import request from 'supertest';
import { App } from 'supertest/types';
import { ThrottlerGuard } from '@nestjs/throttler';
import { AppModule } from '../src/app.module';
import { PrismaService } from '../src/prisma/prisma.service';
import cookieParser from 'cookie-parser';

// Mock bcrypt
jest.mock('bcrypt', () => ({
  hash: jest.fn().mockResolvedValue('hashed_password'),
  compare: jest.fn(),
}));

// eslint-disable-next-line @typescript-eslint/no-require-imports
const bcrypt = require('bcrypt');

describe('Auth (e2e)', () => {
  let app: INestApplication<App>;

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
    // Disable rate limiting for e2e tests
    jest
      .spyOn(ThrottlerGuard.prototype, 'canActivate')
      .mockResolvedValue(true);

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
  });

  afterAll(async () => {
    await app.close();
  });

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('POST /auth/register', () => {
    it('should register a user and set cookies', async () => {
      mockPrisma.user.create.mockResolvedValue({
        id: 'new-user-uuid',
        username: 'newuser',
        role: 'user',
      });
      mockPrisma.refreshToken.create.mockResolvedValue({});

      const res = await request(app.getHttpServer())
        .post('/auth/register')
        .send({
          phone: '0612345678',
          username: 'newuser',
          password: 'Password1!',
          email: 'new@example.com',
        })
        .expect(201);

      expect(res.body).toEqual({ message: 'registered' });

      const cookies = res.headers['set-cookie'];
      expect(cookies).toBeDefined();
      const cookieStr = Array.isArray(cookies) ? cookies.join('; ') : cookies;
      expect(cookieStr).toContain('ACCESS_TOKEN');
      expect(cookieStr).toContain('REFRESH_TOKEN');
    });

    it('should return 400 for missing required fields', async () => {
      await request(app.getHttpServer())
        .post('/auth/register')
        .send({ phone: '0612345678' })
        .expect(400);
    });

    it('should return 400 for weak password', async () => {
      await request(app.getHttpServer())
        .post('/auth/register')
        .send({
          phone: '0612345678',
          username: 'newuser',
          password: 'short',
        })
        .expect(400);
    });

    it('should return 400 for password without special char', async () => {
      await request(app.getHttpServer())
        .post('/auth/register')
        .send({
          phone: '0612345678',
          username: 'newuser',
          password: 'Password123',
        })
        .expect(400);
    });

    it('should return 409 if username already taken', async () => {
      const { Prisma } = jest.requireActual('@prisma/client');
      mockPrisma.user.create.mockRejectedValue(
        new Prisma.PrismaClientKnownRequestError('Unique constraint failed', {
          code: 'P2002',
          clientVersion: '5.0.0',
          meta: { target: ['username'] },
        }),
      );

      await request(app.getHttpServer())
        .post('/auth/register')
        .send({
          phone: '0699999999',
          username: 'taken',
          password: 'Password1!',
        })
        .expect(409);
    });
  });

  describe('POST /auth/login', () => {
    it('should login and set cookies', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'testuser',
        passwordHash: 'hashed_password',
        role: 'user',
      });
      (bcrypt.compare as jest.Mock).mockResolvedValue(true);
      mockPrisma.refreshToken.create.mockResolvedValue({});

      const res = await request(app.getHttpServer())
        .post('/auth/login')
        .send({ username: 'testuser', password: 'Password1!' })
        .expect(201);

      expect(res.body).toEqual({ message: 'logged in' });

      const cookies = res.headers['set-cookie'];
      expect(cookies).toBeDefined();
    });

    it('should return 401 for invalid credentials', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await request(app.getHttpServer())
        .post('/auth/login')
        .send({ username: 'nouser', password: 'Password1!' })
        .expect(401);
    });

    it('should return 401 for wrong password', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'testuser',
        passwordHash: 'hashed_password',
        role: 'user',
      });
      (bcrypt.compare as jest.Mock).mockResolvedValue(false);

      await request(app.getHttpServer())
        .post('/auth/login')
        .send({ username: 'testuser', password: 'Wrong1!' })
        .expect(401);
    });

    it('should return 400 for missing fields', async () => {
      await request(app.getHttpServer())
        .post('/auth/login')
        .send({})
        .expect(400);
    });
  });

  describe('POST /auth/refresh', () => {
    it('should refresh token when valid refresh cookie is provided', async () => {
      mockPrisma.refreshToken.findUnique.mockResolvedValue({
        id: 'token-id',
        token: 'valid-refresh',
        expiresAt: new Date(Date.now() + 86400000),
        user: { id: 'user-uuid', username: 'testuser', role: 'user' },
      });

      const res = await request(app.getHttpServer())
        .post('/auth/refresh')
        .set('Cookie', ['REFRESH_TOKEN=valid-refresh'])
        .expect(201);

      expect(res.body).toEqual({ message: 'token refreshed' });
    });

    it('should return 401 when no refresh token cookie', async () => {
      await request(app.getHttpServer())
        .post('/auth/refresh')
        .expect(401);
    });
  });

  describe('POST /auth/logout', () => {
    it('should logout and clear cookies', async () => {
      mockPrisma.refreshToken.deleteMany.mockResolvedValue({ count: 1 });

      const res = await request(app.getHttpServer())
        .post('/auth/logout')
        .set('Cookie', ['REFRESH_TOKEN=some-token'])
        .expect(201);

      expect(res.body).toEqual({ message: 'logged out' });
    });
  });

  describe('GET /auth/me', () => {
    it('should return 401 without auth cookie', async () => {
      await request(app.getHttpServer())
        .get('/auth/me')
        .expect(401);
    });
  });
});
