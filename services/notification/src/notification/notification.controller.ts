import {
  Controller,
  Post,
  Get,
  Patch,
  Delete,
  Body,
  Param,
  Query,
  UseGuards,
  Request,
  ParseUUIDPipe,
  ForbiddenException,
} from '@nestjs/common';
import { NotificationType } from '@prisma/client';
import { NotificationService } from './notification.service';
import { CreateNotificationDto } from './dto/create-notification.dto';
import { UpdateSettingsDto } from './dto/update-settings.dto';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

interface AuthRequest {
  user: { userId: string; username: string };
}

@UseGuards(JwtAuthGuard)
@Controller('notifications')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Post()
  async create(@Body() dto: CreateNotificationDto) {
    return this.notificationService.create(dto);
  }

  @Get('settings/:userId')
  async getSettings(
    @Request() req: AuthRequest,
    @Param('userId', ParseUUIDPipe) userId: string,
  ) {
    if (req.user.userId !== userId) {
      throw new ForbiddenException('Accès non autorisé');
    }
    return this.notificationService.getSettings(userId);
  }

  @Patch('settings/:userId')
  async updateSettings(
    @Request() req: AuthRequest,
    @Param('userId', ParseUUIDPipe) userId: string,
    @Body() dto: UpdateSettingsDto,
  ) {
    if (req.user.userId !== userId) {
      throw new ForbiddenException('Accès non autorisé');
    }
    return this.notificationService.updateSettings(userId, dto);
  }

  @Get(':userId/unread-count')
  async getUnreadCount(
    @Request() req: AuthRequest,
    @Param('userId', ParseUUIDPipe) userId: string,
  ) {
    if (req.user.userId !== userId) {
      throw new ForbiddenException('Accès non autorisé');
    }
    return this.notificationService.getUnreadCount(userId);
  }

  @Get(':userId')
  async findByUser(
    @Request() req: AuthRequest,
    @Param('userId', ParseUUIDPipe) userId: string,
    @Query('page') page?: string,
    @Query('limit') limit?: string,
    @Query('type') type?: NotificationType,
    @Query('isRead') isRead?: string,
  ) {
    if (req.user.userId !== userId) {
      throw new ForbiddenException('Accès non autorisé');
    }

    const pageNum = Math.max(1, Number(page) || 1);
    const limitNum = Math.min(100, Math.max(1, Number(limit) || 20));

    return this.notificationService.findByUser(
      userId,
      pageNum,
      limitNum,
      type,
      isRead !== undefined ? isRead === 'true' : undefined,
    );
  }

  @Patch(':id/read')
  async markAsRead(@Param('id', ParseUUIDPipe) id: string) {
    return this.notificationService.markAsRead(id);
  }

  @Patch(':userId/read-all')
  async markAllAsRead(
    @Request() req: AuthRequest,
    @Param('userId', ParseUUIDPipe) userId: string,
  ) {
    if (req.user.userId !== userId) {
      throw new ForbiddenException('Accès non autorisé');
    }
    return this.notificationService.markAllAsRead(userId);
  }

  @Delete(':id')
  async remove(@Param('id', ParseUUIDPipe) id: string) {
    return this.notificationService.remove(id);
  }
}
