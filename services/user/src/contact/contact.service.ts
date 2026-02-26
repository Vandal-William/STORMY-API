import {
  BadRequestException,
  ConflictException,
  Injectable,
  NotFoundException,
} from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { PrismaService } from '../prisma/prisma.service';
import {
  AddContactDto,
  BlockUserDto,
  UpdateContactDto,
} from './dto/contact.dto';

const CONTACT_USER_SELECT = {
  id: true,
  username: true,
  avatarUrl: true,
  about: true,
  lastSeen: true,
} as const;

@Injectable()
export class ContactService {
  constructor(private readonly prisma: PrismaService) {}

  // ========================
  // CONTACTS
  // ========================

  async addContact(userId: string, dto: AddContactDto) {
    if (userId === dto.contactUserId) {
      throw new BadRequestException(
        'Vous ne pouvez pas vous ajouter vous-même en contact',
      );
    }

    const targetUser = await this.prisma.user.findUnique({
      where: { id: dto.contactUserId },
    });

    if (!targetUser) {
      throw new NotFoundException('Utilisateur non trouvé');
    }

    const isBlocked = await this.prisma.blockedUser.findFirst({
      where: {
        OR: [
          { userId, blockedUserId: dto.contactUserId },
          { userId: dto.contactUserId, blockedUserId: userId },
        ],
      },
    });

    if (isBlocked) {
      throw new BadRequestException(
        'Impossible d\'ajouter un utilisateur bloqué en contact',
      );
    }

    try {
      return await this.prisma.contact.create({
        data: {
          userId,
          contactUserId: dto.contactUserId,
          nickname: dto.nickname,
        },
        include: {
          contactUser: { select: CONTACT_USER_SELECT },
        },
      });
    } catch (error) {
      if (
        error instanceof Prisma.PrismaClientKnownRequestError &&
        error.code === 'P2002'
      ) {
        throw new ConflictException('Ce contact existe déjà');
      }
      throw error;
    }
  }

  async getContacts(userId: string, page: number, limit: number) {
    const where = { userId };

    const [data, total] = await Promise.all([
      this.prisma.contact.findMany({
        where,
        include: {
          contactUser: { select: CONTACT_USER_SELECT },
        },
        orderBy: { createdAt: 'desc' },
        skip: (page - 1) * limit,
        take: limit,
      }),
      this.prisma.contact.count({ where }),
    ]);

    return {
      data,
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    };
  }

  async updateContact(userId: string, contactId: string, dto: UpdateContactDto) {
    const contact = await this.prisma.contact.findUnique({
      where: { id: contactId },
    });

    if (!contact || contact.userId !== userId) {
      throw new NotFoundException('Contact non trouvé');
    }

    return this.prisma.contact.update({
      where: { id: contactId },
      data: { nickname: dto.nickname },
      include: {
        contactUser: { select: CONTACT_USER_SELECT },
      },
    });
  }

  async removeContact(userId: string, contactId: string) {
    const contact = await this.prisma.contact.findUnique({
      where: { id: contactId },
    });

    if (!contact || contact.userId !== userId) {
      throw new NotFoundException('Contact non trouvé');
    }

    await this.prisma.contact.delete({ where: { id: contactId } });

    return { deleted: true };
  }

  // ========================
  // BLOCKED USERS
  // ========================

  async blockUser(userId: string, dto: BlockUserDto) {
    if (userId === dto.blockedUserId) {
      throw new BadRequestException(
        'Vous ne pouvez pas vous bloquer vous-même',
      );
    }

    const targetUser = await this.prisma.user.findUnique({
      where: { id: dto.blockedUserId },
    });

    if (!targetUser) {
      throw new NotFoundException('Utilisateur non trouvé');
    }

    try {
      const [blocked] = await this.prisma.$transaction([
        this.prisma.blockedUser.create({
          data: {
            userId,
            blockedUserId: dto.blockedUserId,
          },
          include: {
            blockedUser: { select: CONTACT_USER_SELECT },
          },
        }),
        this.prisma.contact.deleteMany({
          where: {
            OR: [
              { userId, contactUserId: dto.blockedUserId },
              { userId: dto.blockedUserId, contactUserId: userId },
            ],
          },
        }),
      ]);

      return blocked;
    } catch (error) {
      if (
        error instanceof Prisma.PrismaClientKnownRequestError &&
        error.code === 'P2002'
      ) {
        throw new ConflictException('Cet utilisateur est déjà bloqué');
      }
      throw error;
    }
  }

  async getBlockedUsers(userId: string, page: number, limit: number) {
    const where = { userId };

    const [data, total] = await Promise.all([
      this.prisma.blockedUser.findMany({
        where,
        include: {
          blockedUser: { select: CONTACT_USER_SELECT },
        },
        orderBy: { createdAt: 'desc' },
        skip: (page - 1) * limit,
        take: limit,
      }),
      this.prisma.blockedUser.count({ where }),
    ]);

    return {
      data,
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    };
  }

  async unblockUser(userId: string, blockedEntryId: string) {
    const entry = await this.prisma.blockedUser.findUnique({
      where: { id: blockedEntryId },
    });

    if (!entry || entry.userId !== userId) {
      throw new NotFoundException('Blocage non trouvé');
    }

    await this.prisma.blockedUser.delete({ where: { id: blockedEntryId } });

    return { deleted: true };
  }
}
