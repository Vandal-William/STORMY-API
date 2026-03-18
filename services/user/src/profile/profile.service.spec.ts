import { Test, TestingModule } from '@nestjs/testing';
import {
  NotFoundException,
  ConflictException,
} from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { ProfileService } from './profile.service';
import { PrismaService } from '../prisma/prisma.service';

describe('ProfileService', () => {
  let service: ProfileService;

  const mockPrisma = {
    user: {
      findUnique: jest.fn(),
      findMany: jest.fn(),
      count: jest.fn(),
      update: jest.fn(),
      delete: jest.fn(),
    },
    contact: {
      deleteMany: jest.fn(),
    },
    blockedUser: {
      deleteMany: jest.fn(),
    },
    $transaction: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        ProfileService,
        { provide: PrismaService, useValue: mockPrisma },
      ],
    }).compile();

    service = module.get<ProfileService>(ProfileService);

    jest.clearAllMocks();
  });

  describe('getOwnProfile', () => {
    it('should return the user own profile with private fields', async () => {
      const user = {
        id: 'user-uuid',
        phone: '0612345678',
        username: 'testuser',
        email: 'test@example.com',
        avatarUrl: null,
        about: 'Hello',
        lastSeen: new Date(),
        createdAt: new Date(),
      };
      mockPrisma.user.findUnique.mockResolvedValue(user);

      const result = await service.getOwnProfile('user-uuid');

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

      await expect(service.getOwnProfile('nonexistent')).rejects.toThrow(
        NotFoundException,
      );
    });
  });

  describe('getUserById', () => {
    it('should return public profile of a user', async () => {
      const user = {
        id: 'other-uuid',
        username: 'otheruser',
        avatarUrl: null,
        about: 'Bio',
        lastSeen: new Date(),
      };
      mockPrisma.user.findUnique.mockResolvedValue(user);

      const result = await service.getUserById('other-uuid');

      expect(result).toEqual(user);
      expect(mockPrisma.user.findUnique).toHaveBeenCalledWith({
        where: { id: 'other-uuid' },
        select: {
          id: true,
          username: true,
          avatarUrl: true,
          about: true,
          lastSeen: true,
        },
      });
    });

    it('should throw NotFoundException if user not found', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(service.getUserById('nonexistent')).rejects.toThrow(
        NotFoundException,
      );
    });
  });

  describe('searchByUsername', () => {
    it('should return paginated search results', async () => {
      const users = [
        { id: '1', username: 'alice', avatarUrl: null, about: null, lastSeen: new Date() },
        { id: '2', username: 'alex', avatarUrl: null, about: null, lastSeen: new Date() },
      ];
      mockPrisma.user.findMany.mockResolvedValue(users);
      mockPrisma.user.count.mockResolvedValue(2);

      const result = await service.searchByUsername('al', 1, 20);

      expect(result).toEqual({
        data: users,
        total: 2,
        page: 1,
        limit: 20,
        totalPages: 1,
      });
      expect(mockPrisma.user.findMany).toHaveBeenCalledWith({
        where: {
          username: { contains: 'al', mode: Prisma.QueryMode.insensitive },
        },
        select: {
          id: true,
          username: true,
          avatarUrl: true,
          about: true,
          lastSeen: true,
        },
        orderBy: { username: 'asc' },
        skip: 0,
        take: 20,
      });
    });

    it('should handle pagination correctly', async () => {
      mockPrisma.user.findMany.mockResolvedValue([]);
      mockPrisma.user.count.mockResolvedValue(50);

      const result = await service.searchByUsername('test', 3, 10);

      expect(result.totalPages).toBe(5);
      expect(result.page).toBe(3);
      expect(mockPrisma.user.findMany).toHaveBeenCalledWith(
        expect.objectContaining({ skip: 20, take: 10 }),
      );
    });
  });

  describe('updateProfile', () => {
    it('should update and return the user profile', async () => {
      const existingUser = { id: 'user-uuid', username: 'oldname' };
      const updatedUser = {
        id: 'user-uuid',
        phone: '0612345678',
        username: 'newname',
        email: 'test@example.com',
        avatarUrl: null,
        about: null,
        lastSeen: new Date(),
        createdAt: new Date(),
      };
      mockPrisma.user.findUnique.mockResolvedValue(existingUser);
      mockPrisma.user.update.mockResolvedValue(updatedUser);

      const result = await service.updateProfile('user-uuid', {
        username: 'newname',
      });

      expect(result).toEqual(updatedUser);
      expect(mockPrisma.user.update).toHaveBeenCalledWith({
        where: { id: 'user-uuid' },
        data: { username: 'newname' },
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

      await expect(
        service.updateProfile('nonexistent', { username: 'newname' }),
      ).rejects.toThrow(NotFoundException);
    });

    it('should throw ConflictException if username is taken (P2002)', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'oldname',
      });
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        {
          code: 'P2002',
          clientVersion: '5.0.0',
          meta: { target: ['username'] },
        },
      );
      mockPrisma.user.update.mockRejectedValue(prismaError);

      await expect(
        service.updateProfile('user-uuid', { username: 'taken' }),
      ).rejects.toThrow(ConflictException);
      await expect(
        service.updateProfile('user-uuid', { username: 'taken' }),
      ).rejects.toThrow("Ce nom d'utilisateur est déjà pris");
    });

    it('should throw generic ConflictException for other unique constraint violations', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'oldname',
      });
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        {
          code: 'P2002',
          clientVersion: '5.0.0',
          meta: { target: ['email'] },
        },
      );
      mockPrisma.user.update.mockRejectedValue(prismaError);

      await expect(
        service.updateProfile('user-uuid', { email: 'taken@example.com' }),
      ).rejects.toThrow('Une valeur unique est déjà utilisée');
    });
  });

  describe('deleteAccount', () => {
    it('should delete user account and related data', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({
        id: 'user-uuid',
        username: 'testuser',
      });
      mockPrisma.$transaction.mockResolvedValue([
        { count: 2 },
        { count: 1 },
        { id: 'user-uuid' },
      ]);

      const result = await service.deleteAccount('user-uuid');

      expect(result).toEqual({ deleted: true });
      expect(mockPrisma.$transaction).toHaveBeenCalled();
    });

    it('should throw NotFoundException if user not found', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(service.deleteAccount('nonexistent')).rejects.toThrow(
        NotFoundException,
      );
    });
  });
});
