import { Injectable, NotFoundException } from '@nestjs/common';
import { ReportReason, ReportStatus } from '@prisma/client';
import { PrismaService } from '../prisma/prisma.service';
import { CreateReportDto } from './dto/create-report.dto';
import { UpdateReportDto } from './dto/update-report.dto';

@Injectable()
export class ReportService {
  constructor(private readonly prisma: PrismaService) {}

  async create(dto: CreateReportDto) {
    return this.prisma.report.create({ data: dto });
  }

  async findAll(
    page: number,
    limit: number,
    status?: ReportStatus,
    reason?: ReportReason,
  ) {
    const where: { status?: ReportStatus; reason?: ReportReason } = {};
    if (status) where.status = status;
    if (reason) where.reason = reason;

    const [data, total] = await Promise.all([
      this.prisma.report.findMany({
        where,
        orderBy: { createdAt: 'desc' },
        skip: (page - 1) * limit,
        take: limit,
      }),
      this.prisma.report.count({ where }),
    ]);

    return {
      data,
      total,
      page,
      limit,
      totalPages: Math.ceil(total / limit),
    };
  }

  async findOne(id: string) {
    const report = await this.prisma.report.findUnique({ where: { id } });

    if (!report) {
      throw new NotFoundException('Signalement non trouvé');
    }

    return report;
  }

  async updateStatus(id: string, dto: UpdateReportDto) {
    const report = await this.prisma.report.findUnique({ where: { id } });

    if (!report) {
      throw new NotFoundException('Signalement non trouvé');
    }

    return this.prisma.report.update({
      where: { id },
      data: { status: dto.status },
    });
  }

  async remove(id: string) {
    const report = await this.prisma.report.findUnique({ where: { id } });

    if (!report) {
      throw new NotFoundException('Signalement non trouvé');
    }

    await this.prisma.report.delete({ where: { id } });

    return { deleted: true };
  }
}
