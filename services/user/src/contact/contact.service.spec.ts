import { Test, TestingModule } from '@nestjs/testing';
import {
  BadRequestException,
  ConflictException,
  NotFoundException,
} from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { ContactService } from './contact.service';
import { PrismaService } from '../prisma/prisma.service';

describe('ContactService', () => {
  let service: ContactService;

  const mockPrisma = {
    user: {
      findUnique: jest.fn(),
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
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        ContactService,
        { provide: PrismaService, useValue: mockPrisma },
      ],
    }).compile();

    service = module.get<ContactService>(ContactService);

    jest.clearAllMocks();
  });

  // ========================
  // CONTACTS
  // ========================

  describe('addContact', () => {
    const userId = 'user-uuid';
    const dto = { contactUserId: 'contact-uuid', nickname: 'Mon ami' };

    it('should add a contact successfully', async () => {
      const createdContact = {
        id: 'contact-entry-uuid',
        userId,
        contactUserId: dto.contactUserId,
        nickname: dto.nickname,
        contactUser: {
          id: dto.contactUserId,
          username: 'friend',
          avatarUrl: null,
          about: null,
          lastSeen: new Date(),
        },
      };
      mockPrisma.user.findUnique.mockResolvedValue({ id: dto.contactUserId });
      mockPrisma.blockedUser.findFirst.mockResolvedValue(null);
      mockPrisma.contact.create.mockResolvedValue(createdContact);

      const result = await service.addContact(userId, dto);

      expect(result).toEqual(createdContact);
      expect(mockPrisma.contact.create).toHaveBeenCalledWith({
        data: {
          userId,
          contactUserId: dto.contactUserId,
          nickname: dto.nickname,
        },
        include: {
          contactUser: {
            select: {
              id: true,
              username: true,
              avatarUrl: true,
              about: true,
              lastSeen: true,
            },
          },
        },
      });
    });

    it('should throw BadRequestException when adding self as contact', async () => {
      await expect(
        service.addContact(userId, { contactUserId: userId }),
      ).rejects.toThrow(BadRequestException);
      await expect(
        service.addContact(userId, { contactUserId: userId }),
      ).rejects.toThrow(
        'Vous ne pouvez pas vous ajouter vous-même en contact',
      );
    });

    it('should throw NotFoundException if target user does not exist', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(service.addContact(userId, dto)).rejects.toThrow(
        NotFoundException,
      );
    });

    it('should throw BadRequestException if user is blocked', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({ id: dto.contactUserId });
      mockPrisma.blockedUser.findFirst.mockResolvedValue({
        id: 'block-entry',
        userId,
        blockedUserId: dto.contactUserId,
      });

      await expect(service.addContact(userId, dto)).rejects.toThrow(
        BadRequestException,
      );
      await expect(service.addContact(userId, dto)).rejects.toThrow(
        "Impossible d'ajouter un utilisateur bloqué en contact",
      );
    });

    it('should throw ConflictException if contact already exists', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({ id: dto.contactUserId });
      mockPrisma.blockedUser.findFirst.mockResolvedValue(null);
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        { code: 'P2002', clientVersion: '5.0.0', meta: { target: ['userId_contactUserId'] } },
      );
      mockPrisma.contact.create.mockRejectedValue(prismaError);

      await expect(service.addContact(userId, dto)).rejects.toThrow(
        ConflictException,
      );
      await expect(service.addContact(userId, dto)).rejects.toThrow(
        'Ce contact existe déjà',
      );
    });
  });

  describe('getContacts', () => {
    it('should return paginated contacts', async () => {
      const contacts = [
        {
          id: 'c1',
          userId: 'user-uuid',
          contactUserId: 'contact-1',
          contactUser: { id: 'contact-1', username: 'alice' },
        },
      ];
      mockPrisma.contact.findMany.mockResolvedValue(contacts);
      mockPrisma.contact.count.mockResolvedValue(1);

      const result = await service.getContacts('user-uuid', 1, 20);

      expect(result).toEqual({
        data: contacts,
        total: 1,
        page: 1,
        limit: 20,
        totalPages: 1,
      });
    });

    it('should calculate totalPages correctly', async () => {
      mockPrisma.contact.findMany.mockResolvedValue([]);
      mockPrisma.contact.count.mockResolvedValue(45);

      const result = await service.getContacts('user-uuid', 1, 10);

      expect(result.totalPages).toBe(5);
    });
  });

  describe('updateContact', () => {
    it('should update contact nickname', async () => {
      mockPrisma.contact.findUnique.mockResolvedValue({
        id: 'contact-entry-uuid',
        userId: 'user-uuid',
      });
      const updatedContact = {
        id: 'contact-entry-uuid',
        userId: 'user-uuid',
        nickname: 'New nickname',
        contactUser: { id: 'contact-uuid', username: 'friend' },
      };
      mockPrisma.contact.update.mockResolvedValue(updatedContact);

      const result = await service.updateContact('user-uuid', 'contact-entry-uuid', {
        nickname: 'New nickname',
      });

      expect(result).toEqual(updatedContact);
    });

    it('should throw NotFoundException if contact not found', async () => {
      mockPrisma.contact.findUnique.mockResolvedValue(null);

      await expect(
        service.updateContact('user-uuid', 'nonexistent', { nickname: 'test' }),
      ).rejects.toThrow(NotFoundException);
    });

    it('should throw NotFoundException if contact belongs to another user', async () => {
      mockPrisma.contact.findUnique.mockResolvedValue({
        id: 'contact-entry-uuid',
        userId: 'other-user-uuid',
      });

      await expect(
        service.updateContact('user-uuid', 'contact-entry-uuid', {
          nickname: 'test',
        }),
      ).rejects.toThrow(NotFoundException);
    });
  });

  describe('removeContact', () => {
    it('should remove a contact successfully', async () => {
      mockPrisma.contact.findUnique.mockResolvedValue({
        id: 'contact-entry-uuid',
        userId: 'user-uuid',
      });
      mockPrisma.contact.delete.mockResolvedValue({});

      const result = await service.removeContact('user-uuid', 'contact-entry-uuid');

      expect(result).toEqual({ deleted: true });
      expect(mockPrisma.contact.delete).toHaveBeenCalledWith({
        where: { id: 'contact-entry-uuid' },
      });
    });

    it('should throw NotFoundException if contact not found or not owned', async () => {
      mockPrisma.contact.findUnique.mockResolvedValue(null);

      await expect(
        service.removeContact('user-uuid', 'nonexistent'),
      ).rejects.toThrow(NotFoundException);
    });
  });

  // ========================
  // BLOCKED USERS
  // ========================

  describe('blockUser', () => {
    const userId = 'user-uuid';
    const dto = { blockedUserId: 'blocked-uuid' };

    it('should block a user and remove mutual contacts', async () => {
      const blocked = {
        id: 'blocked-entry-uuid',
        userId,
        blockedUserId: dto.blockedUserId,
        blockedUser: {
          id: dto.blockedUserId,
          username: 'blocked',
          avatarUrl: null,
          about: null,
          lastSeen: new Date(),
        },
      };
      mockPrisma.user.findUnique.mockResolvedValue({ id: dto.blockedUserId });
      mockPrisma.$transaction.mockResolvedValue([blocked, { count: 1 }]);

      const result = await service.blockUser(userId, dto);

      expect(result).toEqual(blocked);
    });

    it('should throw BadRequestException when blocking self', async () => {
      await expect(
        service.blockUser(userId, { blockedUserId: userId }),
      ).rejects.toThrow(BadRequestException);
      await expect(
        service.blockUser(userId, { blockedUserId: userId }),
      ).rejects.toThrow('Vous ne pouvez pas vous bloquer vous-même');
    });

    it('should throw NotFoundException if target user does not exist', async () => {
      mockPrisma.user.findUnique.mockResolvedValue(null);

      await expect(service.blockUser(userId, dto)).rejects.toThrow(
        NotFoundException,
      );
    });

    it('should throw ConflictException if user already blocked', async () => {
      mockPrisma.user.findUnique.mockResolvedValue({ id: dto.blockedUserId });
      const prismaError = new Prisma.PrismaClientKnownRequestError(
        'Unique constraint failed',
        { code: 'P2002', clientVersion: '5.0.0', meta: { target: ['userId_blockedUserId'] } },
      );
      mockPrisma.$transaction.mockRejectedValue(prismaError);

      await expect(service.blockUser(userId, dto)).rejects.toThrow(
        ConflictException,
      );
      await expect(service.blockUser(userId, dto)).rejects.toThrow(
        'Cet utilisateur est déjà bloqué',
      );
    });
  });

  describe('getBlockedUsers', () => {
    it('should return paginated blocked users', async () => {
      const blockedUsers = [
        {
          id: 'blocked-entry-1',
          userId: 'user-uuid',
          blockedUserId: 'blocked-1',
          blockedUser: { id: 'blocked-1', username: 'baduser' },
        },
      ];
      mockPrisma.blockedUser.findMany.mockResolvedValue(blockedUsers);
      mockPrisma.blockedUser.count.mockResolvedValue(1);

      const result = await service.getBlockedUsers('user-uuid', 1, 20);

      expect(result).toEqual({
        data: blockedUsers,
        total: 1,
        page: 1,
        limit: 20,
        totalPages: 1,
      });
    });
  });

  describe('unblockUser', () => {
    it('should unblock a user successfully', async () => {
      mockPrisma.blockedUser.findUnique.mockResolvedValue({
        id: 'blocked-entry-uuid',
        userId: 'user-uuid',
        blockedUserId: 'blocked-uuid',
      });
      mockPrisma.blockedUser.delete.mockResolvedValue({});

      const result = await service.unblockUser('user-uuid', 'blocked-entry-uuid');

      expect(result).toEqual({ deleted: true });
    });

    it('should throw NotFoundException if blocked entry not found', async () => {
      mockPrisma.blockedUser.findUnique.mockResolvedValue(null);

      await expect(
        service.unblockUser('user-uuid', 'nonexistent'),
      ).rejects.toThrow(NotFoundException);
    });

    it('should throw NotFoundException if blocked entry belongs to another user', async () => {
      mockPrisma.blockedUser.findUnique.mockResolvedValue({
        id: 'blocked-entry-uuid',
        userId: 'other-user-uuid',
        blockedUserId: 'blocked-uuid',
      });

      await expect(
        service.unblockUser('user-uuid', 'blocked-entry-uuid'),
      ).rejects.toThrow(NotFoundException);
    });
  });
});
