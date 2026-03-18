import { Test, TestingModule } from '@nestjs/testing';
import { JwtService } from '@nestjs/jwt';
import {
  ConflictException,
  NotFoundException,
  UnauthorizedException,
} from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { AuthService } from './auth.service';
import { PrismaService } from '../prisma/prisma.service';

// Mock bcrypt
jest.mock('bcrypt', () => ({
  hash: jest.fn().mockResolvedValue('hashed_password'),
  compare: jest.fn(),
}));

// eslint-disable-next-line @typescript-eslint/no-require-imports
const bcrypt = require('bcrypt');

describe('AuthService', () => {
  let service: AuthService;
  let prisma: PrismaService;
  let jwtService: JwtService;

  const mockPrisma = {
    user: {
      create: jest.fn(),
      findUnique: jest.fn(),
    },
    refreshToken: {
      create: jest.fn(),
      findUnique: jest.fn(),
      delete: jest.fn(),
      deleteMany: jest.fn(),
    },
  };

  const mockJwtService = {
    sign: jest.fn().mockReturnValue('mock_jwt_token'),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        AuthService,
        { provide: PrismaService, useValue: mockPrisma },
        { provide: JwtService, useValue: mockJwtService },
      ],
    }).compile();

    service = module.get<AuthService>(AuthService);
    prisma = module.get<PrismaService>(PrismaService);
    jwtService = module.get<JwtService>(JwtService);

    jest.clearAllMocks();
  });

  describe('register', () => {
    const registerDto = {
      phone: '0612345678',
      username: 'testuser',
      password: 'Password1!',
      email: 'test@example.com',
    };

    it('should register a new user and return tokens', async () => {
      const createdUser = {
        id: 'user-uuid',
        username: 'testuser',
        role: 'user',
      };
      mockPrisma.user.create.mockResolvedValue(createdUser);
      mockPrisma.refreshToken.create.mockResolvedValue({});

      const result = await service.register(registerDto);

      expect(result).toHaveProperty('access_token');
      expect(result).toHaveProperty('refresh_token');
      expect(mockPrisma.user.create).toHaveBeenCalledWith({
        data: {
          phone: registerDto.phone,
          username: registerDto.username,
          passwordHash: 'hashed_password',
          email: registerDto.email,
        },
      });
      expect(mockJwtService.sign).toHaveBeenCalledWith({
        sub: 'user-uuid',
        username: 'testuser',
        role: 'user',
      });
    });

    it('should throw ConflictException if phone is already taken', async () => {
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        { code: 'P2002', clientVersion: '5.0.0', meta: { target: ['phone'] } },
      );
      mockPrisma.user.create.mockRejectedValue(prismaError);

      await expect(service.register(registerDto)).rejects.toThrow(
        ConflictException,
      );
      await expect(service.register(registerDto)).rejects.toThrow(
        'Phone number already taken',
      );
    });

    it('should throw ConflictException if username is already taken', async () => {
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        {
          code: 'P2002',
          clientVersion: '5.0.0',
          meta: { target: ['username'] },
        },
      );
      mockPrisma.user.create.mockRejectedValue(prismaError);

      await expect(service.register(registerDto)).rejects.toThrow(
        ConflictException,
      );
      await expect(service.register(registerDto)).rejects.toThrow(
        'Username already taken',
      );
    });

    it('should throw ConflictException with generic message for other P2002 errors', async () => {
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        { code: 'P2002', clientVersion: '5.0.0', meta: { target: ['email'] } },
      );
      mockPrisma.user.create.mockRejectedValue(prismaError);

      await expect(service.register(registerDto)).rejects.toThrow(
        'Phone number or username already taken',
      );
    });

    it('should rethrow non-Prisma errors', async () => {
      mockPrisma.user.create.mockRejectedValue(new Error('DB connection lost'));

      await expect(service.register(registerDto)).rejects.toThrow(
        'DB connection lost',
      );
    });
  });

  describe('login', () => {
    const loginDto = { username: 'testuser', password: 'Password1!' };

    it('should login and return tokens', async () => {
      const user = {
        id: 'user-uuid',
        username: 'testuser',
        passwordHash: 'hashed_password',
        role: 'user',
      };
      mockPrisma.user.findUnique.mockResolvedValue(user);
      (bcrypt.compare as jest.Mock).mockResolvedValue(true);
      mockPrisma.refreshToken.create.mockResolvedValue({});

      const result = await service.login(loginDto);

      expect(result).toHaveProperty('access_token');
      expect(result).toHaveProperty('refresh_token');
      expect(mockPrisma.user.findUnique).toHaveBeenCalledWith({
        where: { username: 'testuser' },
      });
    });

    it('should throw UnauthorizedException if user not found', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(service.login(loginDto)).rejects.toThrow(
        UnauthorizedException,
      );
      await expect(service.login(loginDto)).rejects.toThrow(
        'Invalid credentials',
      );
    });

    it('should throw UnauthorizedException if password is invalid', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'testuser',
        passwordHash: 'hashed_password',
        role: 'user',
      });
      (bcrypt.compare as jest.Mock).mockResolvedValue(false);

      await expect(service.login(loginDto)).rejects.toThrow(
        UnauthorizedException,
      );
    });
  });

  describe('refreshAccessToken', () => {
    it('should return a new access token for a valid refresh token', async () => {
      const storedToken = {
        id: 'token-id',
        token: 'valid-refresh-token',
        expiresAt: new Date(Date.now() + 86400000), // tomorrow
        user: { id: 'user-uuid', username: 'testuser', role: 'user' },
      };
      mockPrisma.refreshToken.findUnique.mockResolvedValue(storedToken);

      const result = await service.refreshAccessToken('valid-refresh-token');

      expect(result).toHaveProperty('access_token');
      expect(mockJwtService.sign).toHaveBeenCalledWith({
        sub: 'user-uuid',
        username: 'testuser',
        role: 'user',
      });
    });

    it('should throw UnauthorizedException if refresh token not found', async () => {
      mockPrisma.refreshToken.findUnique.mockResolvedValue(null);

      await expect(
        service.refreshAccessToken('invalid-token'),
      ).rejects.toThrow(UnauthorizedException);
      await expect(
        service.refreshAccessToken('invalid-token'),
      ).rejects.toThrow('Invalid refresh token');
    });

    it('should throw UnauthorizedException if refresh token is expired', async () => {
      const storedToken = {
        id: 'token-id',
        token: 'expired-token',
        expiresAt: new Date(Date.now() - 86400000), // yesterday
        user: { id: 'user-uuid', username: 'testuser', role: 'user' },
      };
      mockPrisma.refreshToken.findUnique.mockResolvedValue(storedToken);

      await expect(
        service.refreshAccessToken('expired-token'),
      ).rejects.toThrow(UnauthorizedException);
      await expect(
        service.refreshAccessToken('expired-token'),
      ).rejects.toThrow('Refresh token expired');
      expect(mockPrisma.refreshToken.delete).toHaveBeenCalledWith({
        where: { id: 'token-id' },
      });
    });
  });

  describe('logout', () => {
    it('should delete refresh token if provided', async () => {
      mockPrisma.refreshToken.deleteMany.mockResolvedValue({ count: 1 });

      await service.logout('some-refresh-token');

      expect(mockPrisma.refreshToken.deleteMany).toHaveBeenCalledWith({
        where: { token: 'some-refresh-token' },
      });
    });

    it('should do nothing if no refresh token provided', async () => {
      await service.logout(undefined);

      expect(mockPrisma.refreshToken.deleteMany).not.toHaveBeenCalled();
    });
  });

  describe('getProfile', () => {
    it('should return user profile', async () => {
      const user = {
        id: 'user-uuid',
        phone: '0612345678',
        username: 'testuser',
        email: 'test@example.com',
        avatarUrl: null,
        about: null,
        lastSeen: new Date(),
        createdAt: new Date(),
      };
      mockPrisma.user.findUnique.mockResolvedValue(user);

      const result = await service.getProfile('user-uuid');

      expect(result).toEqual(user);
      expect(mockPrisma.user.findUnique).toHaveBeenCalledWith({
        where: { id: 'user-uuid' },
        select: {
          id: true,
          phone: true,
          username: true,
          email: true,
          avatarUrl: true,
          about: true,
          lastSeen: true,
          createdAt: true,
        },
      });
    });

    it('should throw NotFoundException if user not found', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(service.getProfile('nonexistent')).rejects.toThrow(
        NotFoundException,
      );
    });
  });
});
