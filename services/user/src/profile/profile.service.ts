import {
  Injectable,
  NotFoundException,
  ConflictException,
} from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { PrismaService } from '../prisma/prisma.service';
import { UpdateProfileDto } from './dto/update-profile.dto';

const USER_OWN_SELECT = {
  id: true,
  phone: true,
  username: true,
  email: true,
  avatarUrl: true,
  about: true,
  lastSeen: true,
  createdAt: true,
} as const;

const USER_PUBLIC_SELECT = {
  id: true,
  username: true,
  avatarUrl: true,
  about: true,
  lastSeen: true,
} as const;

@Injectable()
export class ProfileService {
  constructor(private readonly prisma: PrismaService) {}

  async getOwnProfile(userId: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: USER_OWN_SELECT,
    });

    if (!user) {
      throw new NotFoundException('Utilisateur non trouvé');
    }

    return user;
  }

  async getUserById(id: string) {
    const user = await this.prisma.user.findUnique({
      where: { id },
      select: USER_PUBLIC_SELECT,
    });

    if (!user) {
      throw new NotFoundException('Utilisateur non trouvé');
    }

    return user;
  }

  async searchByUsername(username: string, page: number, limit: number) {
    const where = {
      username: {
        contains: username,
        mode: Prisma.QueryMode.insensitive,
      },
    };

    const [data, total] = await Promise.all([
      this.prisma.user.findMany({
        where,
        select: USER_PUBLIC_SELECT,
        orderBy: { username: 'asc' },
        skip: (page - 1) * limit,
        take: limit,
      }),
      this.prisma.user.count({ where }),
    ]);

    return {
      data,
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    };
  }

  async updateProfile(userId: string, dto: UpdateProfileDto) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
    });

    if (!user) {
      throw new NotFoundException('Utilisateur non trouvé');
    }

    try {
      return await this.prisma.user.update({
        where: { id: userId },
        data: dto,
        select: USER_OWN_SELECT,
      });
    } catch (error) {
      if (
        error instanceof Prisma.PrismaClientKnownRequestError &&
        error.code === 'P2002'
      ) {
        const target = error.meta?.target as string[] | undefined;
        if (target?.includes('username')) {
          throw new ConflictException("Ce nom d'utilisateur est déjà pris");
        }
        throw new ConflictException('Une valeur unique est déjà utilisée');
      }
      throw error;
    }
  }

  async deleteAccount(userId: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
    });

    if (!user) {
      throw new NotFoundException('Utilisateur non trouvé');
    }

    await this.prisma.$transaction([
      this.prisma.contact.deleteMany({
        where: { OR: [{ userId }, { contactUserId: userId }] },
      }),
      this.prisma.blockedUser.deleteMany({
        where: { OR: [{ userId }, { blockedUserId: userId }] },
      }),
      this.prisma.user.delete({ where: { id: userId } }),
    ]);

    return { deleted: true };
  }
}
