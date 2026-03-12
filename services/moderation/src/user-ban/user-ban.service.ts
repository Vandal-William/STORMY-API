import {
  ConflictException,
  Injectable,
  NotFoundException,
} from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';
import { CreateBanDto } from './dto/create-ban.dto';
import { UpdateBanDto } from './dto/update-ban.dto';

@Injectable()
export class UserBanService {
  constructor(private readonly prisma: PrismaService) {}

  async create(dto: CreateBanDto) {
    const activeBan = await this.prisma.userBan.findFirst({
      where: {
        userId: dto.userId,
        OR: [{ expiresAt: null }, { expiresAt: { gt: new Date() } }],
      },
    });

    if (activeBan) {
      throw new ConflictException('Cet utilisateur est déjà banni');
    }

    return this.prisma.userBan.create({
      data: {
        userId: dto.userId,
        reason: dto.reason,
        expiresAt: dto.expiresAt ? new Date(dto.expiresAt) : null,
      },
    });
  }

  async findAll(page: number, limit: number, active?: string) {
    const where: Record<string, unknown> = {};

    if (active === 'true') {
      where.OR = [{ expiresAt: null }, { expiresAt: { gt: new Date() } }];
    } else if (active === 'false') {
      where.expiresAt = { not: null, lte: new Date() };
    }

    const [data, total] = await Promise.all([
      this.prisma.userBan.findMany({
        where,
        orderBy: { createdAt: 'desc' },
        skip: (page - 1) * limit,
        take: limit,
      }),
      this.prisma.userBan.count({ where }),
    ]);

    return {
      data,
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    };
  }

  async checkBan(userId: string) {
    const ban = await this.prisma.userBan.findFirst({
      where: {
        userId,
        OR: [{ expiresAt: null }, { expiresAt: { gt: new Date() } }],
      },
      orderBy: { createdAt: 'desc' },
    });

    if (!ban) {
      return { banned: false };
    }

    return { banned: true, ban };
  }

  async update(id: string, dto: UpdateBanDto) {
    const ban = await this.prisma.userBan.findUnique({ where: { id } });

    if (!ban) {
      throw new NotFoundException('Ban non trouvé');
    }

    return this.prisma.userBan.update({
      where: { id },
      data: {
        ...(dto.reason !== undefined && { reason: dto.reason }),
        ...(dto.expiresAt !== undefined && {
          expiresAt: dto.expiresAt ? new Date(dto.expiresAt) : null,
        }),
      },
    });
  }

  async findOne(id: string) {
    const ban = await this.prisma.userBan.findUnique({ where: { id } });

    if (!ban) {
      throw new NotFoundException('Ban non trouvé');
    }

    return ban;
  }

  async remove(id: string) {
    const ban = await this.prisma.userBan.findUnique({ where: { id } });

    if (!ban) {
      throw new NotFoundException('Ban non trouvé');
    }

    await this.prisma.userBan.delete({ where: { id } });

    return { deleted: true };
  }
}
