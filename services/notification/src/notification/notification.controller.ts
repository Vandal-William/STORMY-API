import {
  Controller,
  Post,
  Get,
  Patch,
  Delete,
  Body,
  Param,
  Query,
  ParseUUIDPipe,
} from '@nestjs/common';
import { NotificationType } from '@prisma/client';
import { NotificationService } from './notification.service';
import { CreateNotificationDto } from './dto/create-notification.dto';
import { UpdateSettingsDto } from './dto/update-settings.dto';

@Controller('notifications')
export class NotificationController {
  constructor(private readonly notificationService: NotificationService) {}

  @Post()
  async create(@Body() dto: CreateNotificationDto) {
    return this.notificationService.create(dto);
  }

  @Get('settings/:userId')
  async getSettings(@Param('userId', ParseUUIDPipe) userId: string) {
    return this.notificationService.getSettings(userId);
  }

  @Patch('settings/:userId')
  async updateSettings(
    @Param('userId', ParseUUIDPipe) userId: string,
    @Body() dto: UpdateSettingsDto,
  ) {
    return this.notificationService.updateSettings(userId, dto);
  }

  @Get(':userId/unread-count')
  async getUnreadCount(@Param('userId', ParseUUIDPipe) userId: string) {
    return this.notificationService.getUnreadCount(userId);
  }

  @Get(':userId')
  async findByUser(
    @Param('userId', ParseUUIDPipe) userId: string,
    @Query('page') page?: string,
    @Query('limit') limit?: string,
    @Query('type') type?: NotificationType,
    @Query('isRead') isRead?: string,
  ) {
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
  async markAllAsRead(@Param('userId', ParseUUIDPipe) userId: string) {
    return this.notificationService.markAllAsRead(userId);
  }

  @Delete(':id')
  async remove(@Param('id', ParseUUIDPipe) id: string) {
    return this.notificationService.remove(id);
  }
}
